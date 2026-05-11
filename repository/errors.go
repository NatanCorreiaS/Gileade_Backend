package repository

import (
	"errors"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("registro não encontrado")
var ErrTicketIndisponivel = errors.New("ticket indisponivel")

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
