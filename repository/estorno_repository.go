package repository

import (
	"context"

	model "gileade/gileade_backend/Model"

	"gorm.io/gorm"
)

type EstornoRepository struct {
	db *gorm.DB
}

// NewEstornoRepository instancia o repositorio de estornos.
func NewEstornoRepository(db *gorm.DB) *EstornoRepository {
	return &EstornoRepository{db: db}
}

// Create insere um estorno no banco.
func (r *EstornoRepository) Create(ctx context.Context, estorno *model.Estorno) error {
	return mapGormErr(r.db.WithContext(ctx).Create(estorno).Error)
}

// GetByID busca um estorno pelo ID.
func (r *EstornoRepository) GetByID(ctx context.Context, id uint64) (model.Estorno, error) {
	var estorno model.Estorno
	err := r.db.WithContext(ctx).
		Preload("Pagamento").
		First(&estorno, id).Error
	return estorno, mapGormErr(err)
}

// GetByIDTransacaoEstorno busca um estorno pelo ID de transacao.
func (r *EstornoRepository) GetByIDTransacaoEstorno(ctx context.Context, idTransacao string) (model.Estorno, error) {
	var estorno model.Estorno
	err := r.db.WithContext(ctx).
		Where("id_transacao_estorno = ?", idTransacao).
		First(&estorno).Error
	return estorno, mapGormErr(err)
}

// ListByPagamentoID lista estornos por pagamento.
func (r *EstornoRepository) ListByPagamentoID(ctx context.Context, pagamentoID uint64, limit, offset int) ([]model.Estorno, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var estornos []model.Estorno
	err := r.db.WithContext(ctx).
		Where("pagamento_id = ?", pagamentoID).
		Order("id asc").
		Limit(limit).
		Offset(offset).
		Find(&estornos).Error
	return estornos, mapGormErr(err)
}

// Delete remove um estorno pelo ID.
func (r *EstornoRepository) Delete(ctx context.Context, id uint64) error {
	return mapGormErr(r.db.WithContext(ctx).Delete(&model.Estorno{}, id).Error)
}

// CreateAndMarkTicketReembolsado cria um estorno e marca o TicketUsuario relacionado como Reembolsado.
// A relação é descoberta via Pagamento -> TicketsUsuarioID dentro da transação.
func (r *EstornoRepository) CreateAndMarkTicketReembolsado(ctx context.Context, estorno *model.Estorno) error {
	return mapGormErr(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(estorno).Error; err != nil {
			return err
		}

		var pagamento model.Pagamento
		err := tx.Select("id", "tickets_usuario_id").First(&pagamento, estorno.PagamentoID).Error
		if err != nil {
			return err
		}

		var tu model.TicketUsuario
		if err := tx.Select("id", "ticket_id").First(&tu, pagamento.TicketsUsuarioID).Error; err != nil {
			return err
		}

		res := tx.Model(&model.TicketUsuario{}).
			Where("id = ?", pagamento.TicketsUsuarioID).
			Update("status", model.TicketsStatusReembolsado)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		res = tx.Model(&model.Ticket{}).
			Where("id = ?", tu.TicketID).
			Update("quantidade_disponivel", gorm.Expr("quantidade_disponivel + 1"))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}))
}
