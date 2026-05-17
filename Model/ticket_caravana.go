package model

import "time"

// TicketCaravana representa a tabela "ticket_caravana".

type TicketCaravana struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`

	TicketCompraID uint64       `gorm:"type:bigint;not null;index" json:"id_ticket_compra"`
	TicketCompra   TicketCompra `gorm:"foreignKey:TicketCompraID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"ticket_compra"`

	TicketID uint64 `gorm:"type:bigint;not null;index" json:"id_ticket"`
	Ticket   Ticket `gorm:"foreignKey:TicketID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"ticket"`

	CPFBeneficiados StringArray `gorm:"type:text[]" json:"cpf_beneficiados"`

	DataCriacao     time.Time `gorm:"autoCreateTime" json:"data_criacao"`
	DataAtualizacao time.Time `gorm:"autoUpdateTime" json:"data_atualizacao"`
}

// TableName define o nome da tabela para TicketCaravana.
func (TicketCaravana) TableName() string {
	return "ticket_caravana"
}
