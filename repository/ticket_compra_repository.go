package repository

import (
	"context"

	model "gileade/gileade_backend/Model"

	"gorm.io/gorm"
)

type TicketCompraRepository struct {
	db *gorm.DB
}

// NewTicketCompraRepository instancia o repositorio de tickets por compra.
func NewTicketCompraRepository(db *gorm.DB) *TicketCompraRepository {
	return &TicketCompraRepository{db: db}
}

// WithTx retorna um repositorio com a transacao aplicada.
func (r *TicketCompraRepository) WithTx(tx *gorm.DB) *TicketCompraRepository {
	return &TicketCompraRepository{db: tx}
}

// Create insere um ticket_compra no banco.
func (r *TicketCompraRepository) Create(ctx context.Context, tc *model.TicketCompra) error {
	return mapGormErr(r.db.WithContext(ctx).Create(tc).Error)
}

// CreateWithDetalhes insere ticket_compra e seus detalhes na mesma transacao.
func (r *TicketCompraRepository) CreateWithDetalhes(
	ctx context.Context,
	tc *model.TicketCompra,
	individuais []model.TicketIndividual,
	duos []model.TicketDuo,
	caravanas []model.TicketCaravana,
) error {
	return mapGormErr(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(tc).Error; err != nil {
			return err
		}

		if len(individuais) > 0 {
			for i := range individuais {
				individuais[i].TicketCompraID = tc.ID
				if individuais[i].TicketID == 0 {
					individuais[i].TicketID = tc.TicketID
				}
			}
			if err := tx.Create(&individuais).Error; err != nil {
				return err
			}
		}

		if len(duos) > 0 {
			for i := range duos {
				duos[i].TicketCompraID = tc.ID
				if duos[i].TicketID == 0 {
					duos[i].TicketID = tc.TicketID
				}
			}
			if err := tx.Create(&duos).Error; err != nil {
				return err
			}
		}

		if len(caravanas) > 0 {
			for i := range caravanas {
				caravanas[i].TicketCompraID = tc.ID
				if caravanas[i].TicketID == 0 {
					caravanas[i].TicketID = tc.TicketID
				}
			}
			if err := tx.Create(&caravanas).Error; err != nil {
				return err
			}
		}

		return nil
	}))
}

// GetByID busca um ticket_compra pelo ID.
func (r *TicketCompraRepository) GetByID(ctx context.Context, id uint64) (model.TicketCompra, error) {
	var tc model.TicketCompra
	err := r.db.WithContext(ctx).
		Preload("Usuario").
		Preload("Ticket").
		First(&tc, id).Error
	return tc, mapGormErr(err)
}

// ListByUsuarioID lista tickets por usuario com paginacao.
func (r *TicketCompraRepository) ListByUsuarioID(ctx context.Context, usuarioID uint64, limit, offset int) ([]model.TicketCompra, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var tcs []model.TicketCompra
	err := r.db.WithContext(ctx).
		Where("usuario_id = ?", usuarioID).
		Preload("Ticket").
		Order("id asc").
		Limit(limit).
		Offset(offset).
		Find(&tcs).Error
	return tcs, mapGormErr(err)
}

// ListByStatus lista tickets por status com paginacao.
func (r *TicketCompraRepository) ListByStatus(ctx context.Context, status model.TicketsStatus, limit, offset int) ([]model.TicketCompra, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var tcs []model.TicketCompra
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("id asc").
		Limit(limit).
		Offset(offset).
		Find(&tcs).Error
	return tcs, mapGormErr(err)
}

// UpdateStatus atualiza o status de um ticket_compra.
func (r *TicketCompraRepository) UpdateStatus(ctx context.Context, id uint64, status model.TicketsStatus) error {
	res := r.db.WithContext(ctx).
		Model(&model.TicketCompra{}).
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

// UpdatePreferenceID atualiza o preference_id de um ticket_compra.
func (r *TicketCompraRepository) UpdatePreferenceID(ctx context.Context, id uint64, preferenceID string) error {
	res := r.db.WithContext(ctx).
		Model(&model.TicketCompra{}).
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

// Delete remove um ticket_compra pelo ID.
func (r *TicketCompraRepository) Delete(ctx context.Context, id uint64) error {
	return mapGormErr(r.db.WithContext(ctx).Delete(&model.TicketCompra{}, id).Error)
}
