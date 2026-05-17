package model

import "time"

// TicketCompra representa a tabela "tickets_compra".
// Ela liga Pessoa (id_usuario) e Ticket (id_ticket) e mantém o status.

type TicketCompra struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`

	UsuarioID uint64 `gorm:"type:bigint;not null;index" json:"id_usuario"`
	Usuario   Pessoa `gorm:"foreignKey:UsuarioID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"usuario"`

	Status TicketsStatus `gorm:"type:text;not null" json:"status"`

	PreferenceID string `gorm:"type:text" json:"preference_id"`

	TicketID uint64 `gorm:"type:bigint;not null;index" json:"id_ticket"`
	Ticket   Ticket `gorm:"foreignKey:TicketID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"ticket"`

	Quantidade uint64 `gorm:"type:bigint;not null;default:1" json:"quantidade"`

	DataCriacao     time.Time `gorm:"autoCreateTime" json:"data_criacao"`
	DataAtualizacao time.Time `gorm:"autoUpdateTime" json:"data_atualizacao"`

	Pagamentos []Pagamento `gorm:"foreignKey:TicketCompraID" json:"pagamentos,omitempty"`
}

// TableName define o nome da tabela para TicketCompra.
func (TicketCompra) TableName() string {
	return "tickets_compra"
}
