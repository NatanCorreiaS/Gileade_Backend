package repository

import (
	"context"

	model "gileade/gileade_backend/Model"

	"gorm.io/gorm"
)

type TicketUsuarioRepository struct {
	db *gorm.DB
}

// NewTicketUsuarioRepository instancia o repositorio de tickets por usuario.
func NewTicketUsuarioRepository(db *gorm.DB) *TicketUsuarioRepository {
	return &TicketUsuarioRepository{db: db}
}

// Create insere um ticket_usuario no banco.
func (r *TicketUsuarioRepository) Create(ctx context.Context, tu *model.TicketUsuario) error {
	return mapGormErr(r.db.WithContext(ctx).Create(tu).Error)
}

// GetByID busca um ticket_usuario pelo ID.
func (r *TicketUsuarioRepository) GetByID(ctx context.Context, id uint64) (model.TicketUsuario, error) {
	var tu model.TicketUsuario
	err := r.db.WithContext(ctx).
		Preload("Usuario").
		Preload("Ticket").
		First(&tu, id).Error
	return tu, mapGormErr(err)
}

// ListByUsuarioID lista tickets por usuario com paginacao.
func (r *TicketUsuarioRepository) ListByUsuarioID(ctx context.Context, usuarioID uint64, limit, offset int) ([]model.TicketUsuario, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var tus []model.TicketUsuario
	err := r.db.WithContext(ctx).
		Where("usuario_id = ?", usuarioID).
		Preload("Ticket").
		Order("id asc").
		Limit(limit).
		Offset(offset).
		Find(&tus).Error
	return tus, mapGormErr(err)
}

// ListByStatus lista tickets por status com paginacao.
func (r *TicketUsuarioRepository) ListByStatus(ctx context.Context, status model.TicketsStatus, limit, offset int) ([]model.TicketUsuario, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var tus []model.TicketUsuario
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("id asc").
		Limit(limit).
		Offset(offset).
		Find(&tus).Error
	return tus, mapGormErr(err)
}

// UpdateStatus atualiza o status de um ticket_usuario.
func (r *TicketUsuarioRepository) UpdateStatus(ctx context.Context, id uint64, status model.TicketsStatus) error {
	res := r.db.WithContext(ctx).
		Model(&model.TicketUsuario{}).
		Where("id = ?", id).
		Update("status", status)
	if res.Error != nil {
		return mapGormErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdatePreferenceID atualiza o preference_id de um ticket_usuario.
func (r *TicketUsuarioRepository) UpdatePreferenceID(ctx context.Context, id uint64, preferenceID string) error {
	res := r.db.WithContext(ctx).
		Model(&model.TicketUsuario{}).
		Where("id = ?", id).
		Update("preference_id", preferenceID)
	if res.Error != nil {
		return mapGormErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete remove um ticket_usuario pelo ID.
func (r *TicketUsuarioRepository) Delete(ctx context.Context, id uint64) error {
	return mapGormErr(r.db.WithContext(ctx).Delete(&model.TicketUsuario{}, id).Error)
}
