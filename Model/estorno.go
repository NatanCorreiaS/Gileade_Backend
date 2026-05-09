package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Estorno representa a tabela "estornos".

type Estorno struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`

	PagamentoID uint64    `gorm:"type:bigint;not null;index" json:"id_pagamentos"`
	Pagamento   Pagamento `gorm:"foreignKey:PagamentoID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"pagamento"`

	IDTransacaoEstorno string          `gorm:"type:text;not null;uniqueIndex" json:"id_transacao_estorno"`
	Valor              decimal.Decimal `gorm:"type:numeric(18,2);not null" json:"valor"`
	Motivo             string          `gorm:"type:text" json:"motivo"`

	DataEstorno time.Time `gorm:"type:timestamp;not null" json:"data_estorno"`
}

func (Estorno) TableName() string {
	return "estornos"
}
