package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Ticket representa a tabela "tickets".

type Ticket struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`

	Nome      string          `gorm:"type:text;not null" json:"nome"`
	Descricao string          `gorm:"type:text" json:"descricao"`
	Preco     decimal.Decimal `gorm:"type:numeric(18,2);not null" json:"preco"`

	QuantidadeDisponivel uint64    `gorm:"type:bigint;not null" json:"quantidade_disponivel"`
	DataEvento           time.Time `gorm:"type:date;not null" json:"data_evento"`

	DataCriacao     time.Time `gorm:"autoCreateTime" json:"data_criacao"`
	DataAtualizacao time.Time `gorm:"autoUpdateTime" json:"data_atualizacao"`

	TicketsUsuario []TicketUsuario `gorm:"foreignKey:TicketID" json:"tickets_usuario,omitempty"`
}

// TableName define o nome da tabela para Ticket.
func (Ticket) TableName() string {
	return "tickets"
}
