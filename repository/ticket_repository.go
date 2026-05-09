package repository

import (
	"context"

	model "gileade/gileade_backend/Model"

	"gorm.io/gorm"
)

type TicketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) Create(ctx context.Context, ticket *model.Ticket) error {
	return mapGormErr(r.db.WithContext(ctx).Create(ticket).Error)
}

func (r *TicketRepository) GetByID(ctx context.Context, id uint64) (model.Ticket, error) {
	var ticket model.Ticket
	err := r.db.WithContext(ctx).First(&ticket, id).Error
	return ticket, mapGormErr(err)
}

func (r *TicketRepository) List(ctx context.Context, limit, offset int) ([]model.Ticket, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var tickets []model.Ticket
	err := r.db.WithContext(ctx).
		Order("id asc").
		Limit(limit).
		Offset(offset).
		Find(&tickets).Error
	return tickets, mapGormErr(err)
}

func (r *TicketRepository) Update(ctx context.Context, ticket *model.Ticket) error {
	return mapGormErr(r.db.WithContext(ctx).Save(ticket).Error)
}

func (r *TicketRepository) Delete(ctx context.Context, id uint64) error {
	return mapGormErr(r.db.WithContext(ctx).Delete(&model.Ticket{}, id).Error)
}
