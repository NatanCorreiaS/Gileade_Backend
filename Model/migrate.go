package model

import "gorm.io/gorm"

// AutoMigrate aplica (ou atualiza) o schema no banco usando Gorm.
// Mantém o escopo apenas para as tabelas do fluxograma atual.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Pessoa{},
		&Ticket{},
		&TicketCompra{},
		&TicketIndividual{},
		&TicketDuo{},
		&TicketCaravana{},
		&Pagamento{},
		&Estorno{},
	)
}
