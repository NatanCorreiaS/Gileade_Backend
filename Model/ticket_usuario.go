package model

import "time"

// TicketUsuario representa a tabela "tickets_usuario".
// Ela liga Pessoa (id_usuario) e Ticket (id_ticket) e mantém o status.

type TicketUsuario struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`

	UsuarioID uint64 `gorm:"type:bigint;not null;index" json:"id_usuario"`
	Usuario   Pessoa `gorm:"foreignKey:UsuarioID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"usuario"`

	Status TicketsStatus `gorm:"type:text;not null" json:"status"`

	TicketID uint64 `gorm:"type:bigint;not null;index" json:"id_ticket"`
	Ticket   Ticket `gorm:"foreignKey:TicketID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"ticket"`

	DataCriacao     time.Time `gorm:"autoCreateTime" json:"data_criacao"`
	DataAtualizacao time.Time `gorm:"autoUpdateTime" json:"data_atualizacao"`

	Pagamentos []Pagamento `gorm:"foreignKey:TicketsUsuarioID" json:"pagamentos,omitempty"`
}

// TableName define o nome da tabela para TicketUsuario.
func (TicketUsuario) TableName() string {
	return "tickets_usuario"
}
