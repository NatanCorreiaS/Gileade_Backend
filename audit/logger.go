package audit

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type Logger struct {
	logger *log.Logger
}

var (
	once     sync.Once
	instance *Logger
)

func GetLogger() *Logger {
	once.Do(func() {
		path := os.Getenv("AUDIT_LOG_PATH")
		if path == "" {
			path = "logs/audit.log"
		}

		dir := filepath.Dir(path)
		_ = os.MkdirAll(dir, 0750)

		file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			// Fallback para stderr se nao conseguir abrir arquivo.
			instance = &Logger{logger: log.New(os.Stderr, "AUDIT ", log.Ldate|log.Ltime|log.LUTC)}
			return
		}

		instance = &Logger{logger: log.New(file, "AUDIT ", log.Ldate|log.Ltime|log.LUTC)}
	})

	return instance
}

func (l *Logger) LogUsuarioOperation(nome, cpf, operacao string, sucesso bool, err error) {
	campos := map[string]any{
		"nome": strings.TrimSpace(nome),
		"cpf":  cpf,
	}
	l.LogEvent(operacao, sucesso, campos, err)
}

func (l *Logger) LogEvent(operacao string, sucesso bool, campos map[string]any, err error) {
	status := "sucesso"
	if !sucesso {
		status = "falha"
	}

	msg := fmt.Sprintf("operacao=%s status=%s", operacao, status)

	if len(campos) > 0 {
		keys := make([]string, 0, len(campos))
		for key := range campos {
			if strings.TrimSpace(key) == "" {
				continue
			}
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			val := sanitizeField(key, campos[key])
			if val == "" {
				continue
			}
			msg = msg + " " + key + "=" + val
		}
	}

	if err != nil {
		msg = msg + " erro=" + err.Error()
	}

	l.logger.Println(msg)
}

var cpfDigits = regexp.MustCompile(`\d`)

func maskCPF(cpf string) string {
	digits := strings.Join(cpfDigits.FindAllString(cpf, -1), "")
	if len(digits) < 5 {
		return "***"
	}
	prefix := digits[:3]
	suffix := digits[len(digits)-2:]
	return prefix + "***" + suffix
}

func maskToken(token string) string {
	cleaned := strings.TrimSpace(token)
	if cleaned == "" {
		return ""
	}
	if len(cleaned) <= 2 {
		return "**"
	}
	return strings.Repeat("*", len(cleaned)-2) + cleaned[len(cleaned)-2:]
}

func sanitizeField(key string, val any) string {
	if val == nil {
		return ""
	}

	lowerKey := strings.ToLower(strings.TrimSpace(key))
	strVal := fmt.Sprintf("%v", val)

	switch {
	case lowerKey == "cpf" || strings.HasSuffix(lowerKey, "_cpf"):
		return maskCPF(strVal)
	case strings.Contains(lowerKey, "token") || lowerKey == "authorization":
		return maskToken(strVal)
	case lowerKey == "senha" || strings.Contains(lowerKey, "password"):
		return "***"
	default:
		return strings.TrimSpace(strVal)
	}
}
