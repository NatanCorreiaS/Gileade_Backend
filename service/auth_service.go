package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"sync"
	"time"

	model "gileade/gileade_backend/Model"
	"gileade/gileade_backend/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrCredenciaisInvalidas = errors.New("cpf ou senha invalidos")
	ErrTokenInvalido        = errors.New("token invalido")
	ErrTokenExpirado        = errors.New("token expirado")
	ErrTokenRevogado        = errors.New("token revogado")
)

type Claims struct {
	UsuarioID   uint64 `json:"usuario_id"`
	TipoUsuario string `json:"tipo_usuario"`
	jwt.RegisteredClaims
}

type AuthService struct {
	db           *gorm.DB
	pessoaRepo   *repository.PessoaRepository
	jwtSecret    []byte
	tokenTTL     time.Duration
	blacklist    map[string]time.Time
	blacklistMu  sync.RWMutex
}

var (
	authServiceOnce     sync.Once
	authServiceInstance *AuthService
)

// NewAuthService cria o servico de autenticacao (singleton).
func NewAuthService(db *gorm.DB) *AuthService {
	authServiceOnce.Do(func() {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = generateRandomSecret()
		}

		ttl := 24 * time.Hour
		if ttlStr := os.Getenv("JWT_TTL_HOURS"); ttlStr != "" {
			if parsed, err := time.ParseDuration(ttlStr + "h"); err == nil {
				ttl = parsed
			}
		}

		svc := &AuthService{
			db:         db,
			pessoaRepo: repository.NewPessoaRepository(db),
			jwtSecret:  []byte(secret),
			tokenTTL:   ttl,
			blacklist:  make(map[string]time.Time),
		}

		go svc.cleanupBlacklist()
		authServiceInstance = svc
	})

	return authServiceInstance
}

// HashPassword gera o hash bcrypt de uma senha em texto puro.
func (s *AuthService) HashPassword(senha string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword verifica se a senha em texto puro corresponde ao hash.
func (s *AuthService) CheckPassword(hash, senha string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha))
	return err == nil
}

// Login autentica um usuario por CPF e senha e retorna o token JWT e os dados do usuario.
func (s *AuthService) Login(ctx context.Context, cpf, senha string) (model.Pessoa, string, error) {
	pessoa, err := s.pessoaRepo.GetByCPF(ctx, cpf)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Pessoa{}, "", ErrCredenciaisInvalidas
		}
		return model.Pessoa{}, "", err
	}

	if !s.CheckPassword(pessoa.Senha, senha) {
		return model.Pessoa{}, "", ErrCredenciaisInvalidas
	}

	token, err := s.GenerateToken(pessoa.ID, string(pessoa.TipoUsuario))
	if err != nil {
		return model.Pessoa{}, "", err
	}

	return pessoa, token, nil
}

// GenerateToken gera um token JWT assinado.
func (s *AuthService) GenerateToken(usuarioID uint64, tipoUsuario string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UsuarioID:   usuarioID,
		TipoUsuario: tipoUsuario,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        generateJTI(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken valida e extrai as claims de um token JWT.
func (s *AuthService) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalido
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpirado
		}
		return nil, ErrTokenInvalido
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalido
	}

	return claims, nil
}

// BlacklistToken adiciona um token a lista de revogacao.
func (s *AuthService) BlacklistToken(tokenStr string, claims *Claims) {
	s.blacklistMu.Lock()
	defer s.blacklistMu.Unlock()

	if claims != nil && claims.ID != "" {
		s.blacklist[claims.ID] = time.Now().UTC().Add(s.tokenTTL)
	}
}

// IsTokenBlacklisted verifica se o JTI do token esta na lista de revogacao.
func (s *AuthService) IsTokenBlacklisted(claims *Claims) bool {
	s.blacklistMu.RLock()
	defer s.blacklistMu.RUnlock()

	if claims == nil || claims.ID == "" {
		return false
	}

	expiry, exists := s.blacklist[claims.ID]
	if !exists {
		return false
	}

	return time.Now().UTC().Before(expiry)
}

// Logout invalida o token, adicionando-o a blacklist.
func (s *AuthService) Logout(tokenStr string) error {
	claims, err := s.ValidateToken(tokenStr)
	if err != nil && !errors.Is(err, ErrTokenExpirado) {
		return err
	}

	if claims != nil {
		s.BlacklistToken(tokenStr, claims)
	}

	return nil
}

func (s *AuthService) cleanupBlacklist() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.blacklistMu.Lock()
		now := time.Now().UTC()
		for jti, expiry := range s.blacklist {
			if now.After(expiry) {
				delete(s.blacklist, jti)
			}
		}
		s.blacklistMu.Unlock()
	}
}

func generateJTI() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("falha ao gerar JTI: " + err.Error())
	}
	return hex.EncodeToString(b)
}

func generateRandomSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("falha ao gerar secret: " + err.Error())
	}
	return hex.EncodeToString(b)
}
