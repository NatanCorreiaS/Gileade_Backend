package repository

import (
	"context"

	model "gileade/gileade_backend/Model"

	"gorm.io/gorm"
)

type PessoaRepository struct {
	db *gorm.DB
}

func NewPessoaRepository(db *gorm.DB) *PessoaRepository {
	return &PessoaRepository{db: db}
}

func (r *PessoaRepository) Create(ctx context.Context, pessoa *model.Pessoa) error {
	return mapGormErr(r.db.WithContext(ctx).Create(pessoa).Error)
}

func (r *PessoaRepository) GetByID(ctx context.Context, id uint64) (model.Pessoa, error) {
	var pessoa model.Pessoa
	err := r.db.WithContext(ctx).First(&pessoa, id).Error
	return pessoa, mapGormErr(err)
}

func (r *PessoaRepository) GetByCPF(ctx context.Context, cpf string) (model.Pessoa, error) {
	var pessoa model.Pessoa
	err := r.db.WithContext(ctx).Where("cpf = ?", cpf).First(&pessoa).Error
	return pessoa, mapGormErr(err)
}

func (r *PessoaRepository) List(ctx context.Context, limit, offset int) ([]model.Pessoa, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var pessoas []model.Pessoa
	err := r.db.WithContext(ctx).
		Order("id asc").
		Limit(limit).
		Offset(offset).
		Find(&pessoas).Error
	return pessoas, mapGormErr(err)
}

func (r *PessoaRepository) Update(ctx context.Context, pessoa *model.Pessoa) error {
	// Save atualiza por PK e grava zero-values também.
	return mapGormErr(r.db.WithContext(ctx).Save(pessoa).Error)
}

func (r *PessoaRepository) Delete(ctx context.Context, id uint64) error {
	return mapGormErr(r.db.WithContext(ctx).Delete(&model.Pessoa{}, id).Error)
}
