package repository

import (
	"errors"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("registro não encontrado")

func mapGormErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}
