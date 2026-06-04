package repository

import (
	"errors"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("registro não encontrado")
var ErrTicketIndisponivel = errors.New("ticket indisponivel")
var ErrTipoTicketInvalido = errors.New("tipo de ticket invalido")
var ErrEstornoPrazoExcedido = errors.New("prazo para estorno excedido")
var ErrEstornoSemPermissao = errors.New("sem permissao para estornar este pagamento")

// mapGormErr normaliza erros do Gorm para erros de dominio.
func mapGormErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}
