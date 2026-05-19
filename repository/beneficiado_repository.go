package repository

import (
	"context"

	model "gileade/gileade_backend/Model"

	"gorm.io/gorm"
)

type BeneficiadoRepository struct {
	db *gorm.DB
}

// NewBeneficiadoRepository instancia o repositorio de beneficiados.
func NewBeneficiadoRepository(db *gorm.DB) *BeneficiadoRepository {
	return &BeneficiadoRepository{db: db}
}

// WithTx retorna um repositorio com a transacao aplicada.
func (r *BeneficiadoRepository) WithTx(tx *gorm.DB) *BeneficiadoRepository {
	return &BeneficiadoRepository{db: tx}
}

// FindByCPFs busca beneficiados por lista de CPFs.
func (r *BeneficiadoRepository) FindByCPFs(ctx context.Context, cpfs []string) ([]model.Beneficiado, error) {
	if len(cpfs) == 0 {
		return nil, nil
	}

	var beneficiados []model.Beneficiado
	err := r.db.WithContext(ctx).
		Where("cpf IN ?", cpfs).
		Find(&beneficiados).Error
	return beneficiados, mapGormErr(err)
}

// CreateMany insere beneficiados no banco.
func (r *BeneficiadoRepository) CreateMany(ctx context.Context, beneficiados *[]model.Beneficiado) error {
	if beneficiados == nil || len(*beneficiados) == 0 {
		return nil
	}
	return mapGormErr(r.db.WithContext(ctx).Create(beneficiados).Error)
}
