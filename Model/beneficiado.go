package model

import "time"

// Beneficiado representa a tabela "beneficiados".

type Beneficiado struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`

	Nome string `gorm:"type:text;not null" json:"nome"`
	CPF  string `gorm:"type:text;not null;uniqueIndex" json:"cpf"`

	Idade   int16  `gorm:"type:smallint;not null" json:"idade"`
	Celular string `gorm:"type:text;not null" json:"celular"`
	Igreja  string `gorm:"type:text;not null" json:"igreja"`

	PapelIgreja PapelIgreja `gorm:"type:text;not null" json:"papel_igreja"`
	EstadoCivil EstadoCivil `gorm:"type:text;not null" json:"estado_civil"`

	Email string `gorm:"type:text;not null" json:"email"`
	Sexo  Sexo   `gorm:"type:text;not null" json:"sexo"`

	Cidade   string   `gorm:"type:text;not null" json:"cidade"`
	EstadoUF EstadoUF `gorm:"type:text;not null" json:"estado_uf"`

	Escolaridade Escolaridade `gorm:"type:text;not null" json:"escolaridade"`

	DataCriacao     time.Time `gorm:"autoCreateTime" json:"data_criacao"`
	DataAtualizacao time.Time `gorm:"autoUpdateTime" json:"data_atualizacao"`
}

// TableName define o nome da tabela para Beneficiado.
func (Beneficiado) TableName() string {
	return "beneficiados"
}
