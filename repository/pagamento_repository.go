package repository

import (
	"context"

	model "gileade/gileade_backend/Model"

	"gorm.io/gorm"
)

type PagamentoRepository struct {
	db *gorm.DB
}

// NewPagamentoRepository instancia o repositorio de pagamentos.
func NewPagamentoRepository(db *gorm.DB) *PagamentoRepository {
	return &PagamentoRepository{db: db}
}

// Create insere um pagamento no banco.
func (r *PagamentoRepository) Create(ctx context.Context, pagamento *model.Pagamento) error {
	return mapGormErr(r.db.WithContext(ctx).Create(pagamento).Error)
}

// GetByID busca um pagamento pelo ID.
func (r *PagamentoRepository) GetByID(ctx context.Context, id uint64) (model.Pagamento, error) {
	var pagamento model.Pagamento
	err := r.db.WithContext(ctx).
		Preload("TicketCompra").
		First(&pagamento, id).Error
	return pagamento, mapGormErr(err)
}

// GetByIDTransacao busca um pagamento pelo ID da transacao.
func (r *PagamentoRepository) GetByIDTransacao(ctx context.Context, idTransacao string) (model.Pagamento, error) {
	var pagamento model.Pagamento
	err := r.db.WithContext(ctx).
		Where("id_transacao = ?", idTransacao).
		First(&pagamento).Error
	return pagamento, mapGormErr(err)
}

// ListByTicketCompraID lista pagamentos por ticket_compra.
func (r *PagamentoRepository) ListByTicketCompraID(ctx context.Context, ticketCompraID uint64, limit, offset int) ([]model.Pagamento, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var pagamentos []model.Pagamento
	err := r.db.WithContext(ctx).
		Where("ticket_compra_id = ?", ticketCompraID).
		Order("id asc").
		Limit(limit).
		Offset(offset).
		Find(&pagamentos).Error
	return pagamentos, mapGormErr(err)
}

// Update atualiza um pagamento existente.
func (r *PagamentoRepository) Update(ctx context.Context, pagamento *model.Pagamento) error {
	return mapGormErr(r.db.WithContext(ctx).Save(pagamento).Error)
}

// Delete remove um pagamento pelo ID.
func (r *PagamentoRepository) Delete(ctx context.Context, id uint64) error {
	return mapGormErr(r.db.WithContext(ctx).Delete(&model.Pagamento{}, id).Error)
}

// CreateAndMarkTicketPago cria um pagamento e marca o TicketCompra como Pago na mesma transação.
// Use quando o estado do ticket não pode divergir do registro de pagamento.
func (r *PagamentoRepository) CreateAndMarkTicketPago(ctx context.Context, pagamento *model.Pagamento) error {
	return mapGormErr(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(pagamento).Error; err != nil {
			return err
		}

		var tc model.TicketCompra
		if err := tx.Preload("Ticket").First(&tc, pagamento.TicketCompraID).Error; err != nil {
			return err
		}

		unidadesPorTicket, err := unidadesPorTicket(tc.Ticket.Tipo)
		if err != nil {
			return err
		}
		quantidade := tc.Quantidade
		if quantidade == 0 {
			quantidade = 1
		}
		quantidadeTotal := unidadesPorTicket * quantidade

		res := tx.Model(&model.Ticket{}).
			Where("id = ? AND quantidade_disponivel >= ?", tc.TicketID, quantidadeTotal).
			Update("quantidade_disponivel", gorm.Expr("quantidade_disponivel - ?", quantidadeTotal))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrTicketIndisponivel
		}

		res = tx.Model(&model.TicketCompra{}).
			Where("id = ?", pagamento.TicketCompraID).
			Update("status", model.TicketsStatusPago)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}))
}

func unidadesPorTicket(tipo model.TipoTicket) (uint64, error) {
	switch tipo {
	case model.TipoTicketIndividual, "":
		return 1, nil
	case model.TipoTicketDuo:
		return 2, nil
	case model.TipoTicketCaravana:
		return 10, nil
	default:
		return 0, ErrTipoTicketInvalido
	}
}
