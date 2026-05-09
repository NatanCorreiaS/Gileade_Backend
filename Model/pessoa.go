package model

import "time"

// Pessoa representa a tabela "pessoas".
// Campos e enums seguem o fluxograma.
//
// Importante (segurança): o campo Senha deve armazenar hash (nunca senha em texto puro).

type Pessoa struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`

	Nome        string      `gorm:"type:text;not null" json:"nome"`
	TipoUsuario TipoUsuario `gorm:"type:text;not null" json:"tipo_usuario"`
	Senha       string      `gorm:"type:text;not null" json:"senha"`

	CPF   string `gorm:"type:text;not null;uniqueIndex" json:"cpf"`
	Idade int16  `gorm:"type:smallint" json:"idade"`

	Celular string `gorm:"type:text" json:"celular"`
	Igreja  string `gorm:"type:text" json:"igreja"`

	PapelIgreja PapelIgreja `gorm:"type:text" json:"papel_igreja"`
	EstadoCivil EstadoCivil `gorm:"type:text" json:"estado_civil"`

	Email string `gorm:"type:text" json:"email"`
	Sexo  Sexo   `gorm:"type:text" json:"sexo"`

	Cidade   string   `gorm:"type:text" json:"cidade"`
	EstadoUF EstadoUF `gorm:"type:text" json:"estado_uf"`

	Escolaridade Escolaridade `gorm:"type:text" json:"escolaridade"`

	DataCriacao     time.Time `gorm:"autoCreateTime" json:"data_criacao"`
	DataAtualizacao time.Time `gorm:"autoUpdateTime" json:"data_atualizacao"`

	TicketsUsuario []TicketUsuario `gorm:"foreignKey:UsuarioID" json:"tickets_usuario,omitempty"`
}

func (Pessoa) TableName() string {
	return "pessoas"
}
