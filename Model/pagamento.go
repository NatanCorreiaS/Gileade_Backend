package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Pagamento representa a tabela "pagamentos".

type Pagamento struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`

	IDTransacao string          `gorm:"type:text;not null;uniqueIndex" json:"id_transacao"`
	Valor       decimal.Decimal `gorm:"type:numeric(18,2);not null" json:"valor"`

	TicketCompraID uint64       `gorm:"type:bigint;not null;index" json:"id_ticket_compra"`
	TicketCompra   TicketCompra `gorm:"foreignKey:TicketCompraID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"ticket_compra"`

	Metodo        MetodoPagamento `gorm:"type:text;not null" json:"metodo"`
	DataPagamento time.Time       `gorm:"type:timestamp;not null" json:"data_pagamento"`

	Estornos []Estorno `gorm:"foreignKey:PagamentoID" json:"estornos,omitempty"`
}

// TableName define o nome da tabela para Pagamento.
func (Pagamento) TableName() string {
	return "pagamentos"
}
