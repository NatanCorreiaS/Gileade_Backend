package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	model "gileade/gileade_backend/Model"
	"gileade/gileade_backend/internal/testutil"
	"gileade/gileade_backend/repository"

	"github.com/shopspring/decimal"
)

// TestPessoaRepository cobre operacoes basicas de pessoa.
func TestPessoaRepository(t *testing.T) {
	tdb := testutil.StartPostgres(t)
	if err := model.AutoMigrate(tdb.DB); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}

	repo := repository.NewPessoaRepository(tdb.DB)
	ctx := context.Background()

	cases := []struct {
		name  string
		cpf   string
		email string
	}{
		{name: "pessoa-1", cpf: "00000000001", email: "fulano1@example.com"},
		{name: "pessoa-2", cpf: "00000000002", email: "fulano2@example.com"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pessoa := model.Pessoa{
				Nome:        "Fulano",
				TipoUsuario: model.TipoUsuarioUsuario,
				Senha:       "hash", // hash fake para teste
				CPF:         tc.cpf,
				Idade:       20,
				Email:       tc.email,
				Sexo:        model.SexoMasculino,
				EstadoUF:    model.EstadoUFSaoPaulo,
				EstadoCivil: model.EstadoCivilSolteiro,
			}

			if err := repo.Create(ctx, &pessoa); err != nil {
				t.Fatalf("Create: %v", err)
			}
			if pessoa.ID == 0 {
				t.Fatalf("expected ID")
			}

			gotByID, err := repo.GetByID(ctx, pessoa.ID)
			if err != nil {
				t.Fatalf("GetByID: %v", err)
			}
			if gotByID.CPF != pessoa.CPF {
				t.Fatalf("cpf mismatch")
			}

			gotByCPF, err := repo.GetByCPF(ctx, pessoa.CPF)
			if err != nil {
				t.Fatalf("GetByCPF: %v", err)
			}
			if gotByCPF.ID != pessoa.ID {
				t.Fatalf("id mismatch")
			}

			list, err := repo.List(ctx, 10, 0)
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if len(list) == 0 {
				t.Fatalf("expected list")
			}

			pessoa.Nome = "Fulano 2"
			if err := repo.Update(ctx, &pessoa); err != nil {
				t.Fatalf("Update: %v", err)
			}
			gotUpdated, err := repo.GetByID(ctx, pessoa.ID)
			if err != nil {
				t.Fatalf("GetByID after update: %v", err)
			}
			if gotUpdated.Nome != "Fulano 2" {
				t.Fatalf("update not persisted")
			}

			if err := repo.Delete(ctx, pessoa.ID); err != nil {
				t.Fatalf("Delete: %v", err)
			}
			_, err = repo.GetByID(ctx, pessoa.ID)
			if !errors.Is(err, repository.ErrNotFound) {
				t.Fatalf("expected ErrNotFound, got %v", err)
			}
		})
	}
}

// TestTicketRepository cobre operacoes basicas de ticket.
func TestTicketRepository(t *testing.T) {
	tdb := testutil.StartPostgres(t)
	if err := model.AutoMigrate(tdb.DB); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}

	repo := repository.NewTicketRepository(tdb.DB)
	ctx := context.Background()

	cases := []struct {
		name string
		nome string
	}{
		{name: "ticket-1", nome: "Ingresso"},
		{name: "ticket-2", nome: "Ingresso 2"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ticket := model.Ticket{
				Tipo:                 model.TipoTicketIndividual,
				Nome:                 tc.nome,
				Descricao:            "VIP",
				Preco:                decimal.NewFromFloat(10.50),
				QuantidadeDisponivel: 100,
				DataEvento:           time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
			}

			if err := repo.Create(ctx, &ticket); err != nil {
				t.Fatalf("Create: %v", err)
			}

			got, err := repo.GetByID(ctx, ticket.ID)
			if err != nil {
				t.Fatalf("GetByID: %v", err)
			}
			if got.Nome != ticket.Nome {
				t.Fatalf("nome mismatch")
			}

			list, err := repo.List(ctx, 10, 0)
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if len(list) == 0 {
				t.Fatalf("expected list")
			}

			ticket.Descricao = "VIP 2"
			if err := repo.Update(ctx, &ticket); err != nil {
				t.Fatalf("Update: %v", err)
			}
			gotUpdated, err := repo.GetByID(ctx, ticket.ID)
			if err != nil {
				t.Fatalf("GetByID after update: %v", err)
			}
			if gotUpdated.Descricao != "VIP 2" {
				t.Fatalf("update not persisted")
			}

			if err := repo.Delete(ctx, ticket.ID); err != nil {
				t.Fatalf("Delete: %v", err)
			}
			_, err = repo.GetByID(ctx, ticket.ID)
			if !errors.Is(err, repository.ErrNotFound) {
				t.Fatalf("expected ErrNotFound, got %v", err)
			}
		})
	}
}

// TestTicketCompraRepository cobre operacoes basicas de ticket_compra.
func TestTicketCompraRepository(t *testing.T) {
	tdb := testutil.StartPostgres(t)
	if err := model.AutoMigrate(tdb.DB); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}

	ctx := context.Background()

	pRepo := repository.NewPessoaRepository(tdb.DB)
	tRepo := repository.NewTicketRepository(tdb.DB)
	tcRepo := repository.NewTicketCompraRepository(tdb.DB)

	cases := []struct {
		name  string
		cpf   string
		email string
	}{
		{name: "tu-1", cpf: "00000000100", email: "beltrano1@example.com"},
		{name: "tu-2", cpf: "00000000101", email: "beltrano2@example.com"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pessoa := model.Pessoa{
				Nome:        "Beltrano",
				TipoUsuario: model.TipoUsuarioUsuario,
				Senha:       "hash",
				CPF:         tc.cpf,
				Idade:       21,
				Email:       tc.email,
				Sexo:        model.SexoMasculino,
				EstadoUF:    model.EstadoUFRioDeJaneiro,
				EstadoCivil: model.EstadoCivilSolteiro,
			}
			if err := pRepo.Create(ctx, &pessoa); err != nil {
				t.Fatalf("seed pessoa: %v", err)
			}

			ticket := model.Ticket{
				Tipo:                 model.TipoTicketIndividual,
				Nome:                 "Ingresso 2",
				Descricao:            "Standard",
				Preco:                decimal.NewFromFloat(5.00),
				QuantidadeDisponivel: 10,
				DataEvento:           time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
			}
			if err := tRepo.Create(ctx, &ticket); err != nil {
				t.Fatalf("seed ticket: %v", err)
			}

			tc := model.TicketCompra{
				UsuarioID:  pessoa.ID,
				TicketID:   ticket.ID,
				Quantidade: 1,
				Status:     model.TicketsStatusPendente,
			}
			if err := tcRepo.Create(ctx, &tc); err != nil {
				t.Fatalf("Create: %v", err)
			}

			got, err := tcRepo.GetByID(ctx, tc.ID)
			if err != nil {
				t.Fatalf("GetByID: %v", err)
			}
			if got.Usuario.ID != pessoa.ID || got.Ticket.ID != ticket.ID {
				t.Fatalf("preload mismatch")
			}

			list, err := tcRepo.ListByUsuarioID(ctx, pessoa.ID, 10, 0)
			if err != nil {
				t.Fatalf("ListByUsuarioID: %v", err)
			}
			if len(list) != 1 {
				t.Fatalf("expected 1 record, got %d", len(list))
			}

			if err := tcRepo.UpdateStatus(ctx, tc.ID, model.TicketsStatusPago); err != nil {
				t.Fatalf("UpdateStatus: %v", err)
			}
			got2, err := tcRepo.GetByID(ctx, tc.ID)
			if err != nil {
				t.Fatalf("GetByID after UpdateStatus: %v", err)
			}
			if got2.Status != model.TicketsStatusPago {
				t.Fatalf("status not updated")
			}

			if err := tcRepo.Delete(ctx, tc.ID); err != nil {
				t.Fatalf("Delete: %v", err)
			}
			_, err = tcRepo.GetByID(ctx, tc.ID)
			if !errors.Is(err, repository.ErrNotFound) {
				t.Fatalf("expected ErrNotFound, got %v", err)
			}
		})
	}
}

// TestPagamentoAndEstornoRepositories valida fluxo de pagamento e estorno.
func TestPagamentoAndEstornoRepositories(t *testing.T) {
	tdb := testutil.StartPostgres(t)
	if err := model.AutoMigrate(tdb.DB); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}

	ctx := context.Background()

	pRepo := repository.NewPessoaRepository(tdb.DB)
	tRepo := repository.NewTicketRepository(tdb.DB)
	tcRepo := repository.NewTicketCompraRepository(tdb.DB)
	payRepo := repository.NewPagamentoRepository(tdb.DB)
	estRepo := repository.NewEstornoRepository(tdb.DB)

	cases := []struct {
		name       string
		txID       string
		estornoTx  string
		metodoNovo model.MetodoPagamento
		cpf        string
		email      string
	}{
		{name: "fluxo-1", txID: "tx-001", estornoTx: "st-001", metodoNovo: model.MetodoPagamentoBoleto, cpf: "00000000200", email: "ciclano1@example.com"},
		{name: "fluxo-2", txID: "tx-002", estornoTx: "st-002", metodoNovo: model.MetodoPagamentoCartaoCredito, cpf: "00000000201", email: "ciclano2@example.com"},
	}

	for _, tcCase := range cases {
		t.Run(tcCase.name, func(t *testing.T) {
			pessoa := model.Pessoa{
				Nome:        "Ciclano",
				TipoUsuario: model.TipoUsuarioUsuario,
				Senha:       "hash",
				CPF:         tcCase.cpf,
				Idade:       22,
				Email:       tcCase.email,
				Sexo:        model.SexoMasculino,
				EstadoUF:    model.EstadoUFMinasGerais,
				EstadoCivil: model.EstadoCivilSolteiro,
			}
			if err := pRepo.Create(ctx, &pessoa); err != nil {
				t.Fatalf("seed pessoa: %v", err)
			}

			ticket := model.Ticket{
				Tipo:                 model.TipoTicketDuo,
				Nome:                 "Ingresso 3",
				Descricao:            "Standard",
				Preco:                decimal.NewFromFloat(15.00),
				QuantidadeDisponivel: 10,
				DataEvento:           time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
			}
			if err := tRepo.Create(ctx, &ticket); err != nil {
				t.Fatalf("seed ticket: %v", err)
			}
			quantidadeInicial := ticket.QuantidadeDisponivel

			compra := model.TicketCompra{
				UsuarioID:  pessoa.ID,
				TicketID:   ticket.ID,
				Quantidade: 2,
				Status:     model.TicketsStatusPendente,
			}
			if err := tcRepo.Create(ctx, &compra); err != nil {
				t.Fatalf("seed ticket compra: %v", err)
			}

			pagamento := model.Pagamento{
				IDTransacao:    tcCase.txID,
				Valor:          decimal.NewFromFloat(15.00),
				TicketCompraID: compra.ID,
				Metodo:         model.MetodoPagamentoPix,
				DataPagamento:  time.Now().UTC(),
			}
			if err := payRepo.CreateAndMarkTicketPago(ctx, &pagamento); err != nil {
				t.Fatalf("CreateAndMarkTicketPago: %v", err)
			}

			gotTicketPago, err := tRepo.GetByID(ctx, ticket.ID)
			if err != nil {
				t.Fatalf("GetByID ticket pago: %v", err)
			}
			if gotTicketPago.QuantidadeDisponivel != quantidadeInicial-4 {
				t.Fatalf("quantidade_disponivel nao atualizada no pagamento")
			}

			gotTC, err := tcRepo.GetByID(ctx, compra.ID)
			if err != nil {
				t.Fatalf("GetByID tc: %v", err)
			}
			if gotTC.Status != model.TicketsStatusPago {
				t.Fatalf("expected status Pago")
			}

			gotPayByTx, err := payRepo.GetByIDTransacao(ctx, tcCase.txID)
			if err != nil {
				t.Fatalf("GetByIDTransacao: %v", err)
			}
			if gotPayByTx.ID == 0 {
				t.Fatalf("expected pagamento ID")
			}

			listPay, err := payRepo.ListByTicketCompraID(ctx, compra.ID, 10, 0)
			if err != nil {
				t.Fatalf("ListByTicketCompraID: %v", err)
			}
			if len(listPay) != 1 {
				t.Fatalf("expected 1 pagamento")
			}

			gotPay, err := payRepo.GetByID(ctx, pagamento.ID)
			if err != nil {
				t.Fatalf("GetByID pagamento: %v", err)
			}
			if gotPay.TicketCompraID != compra.ID {
				t.Fatalf("ticket compra mismatch")
			}

			estorno := model.Estorno{
				PagamentoID:        pagamento.ID,
				IDTransacaoEstorno: tcCase.estornoTx,
				Valor:              decimal.NewFromFloat(15.00),
				Motivo:             "teste",
				DataEstorno:        time.Now().UTC(),
			}
			if err := estRepo.CreateAndMarkTicketReembolsado(ctx, &estorno); err != nil {
				t.Fatalf("CreateAndMarkTicketReembolsado: %v", err)
			}

			gotTicketEstorno, err := tRepo.GetByID(ctx, ticket.ID)
			if err != nil {
				t.Fatalf("GetByID ticket estorno: %v", err)
			}
			if gotTicketEstorno.QuantidadeDisponivel != quantidadeInicial {
				t.Fatalf("quantidade_disponivel nao restaurada no estorno")
			}

			gotTC2, err := tcRepo.GetByID(ctx, compra.ID)
			if err != nil {
				t.Fatalf("GetByID tc after estorno: %v", err)
			}
			if gotTC2.Status != model.TicketsStatusReembolsado {
				t.Fatalf("expected status Reembolsado")
			}

			gotEstTx, err := estRepo.GetByIDTransacaoEstorno(ctx, tcCase.estornoTx)
			if err != nil {
				t.Fatalf("GetByIDTransacaoEstorno: %v", err)
			}
			if gotEstTx.ID == 0 {
				t.Fatalf("expected estorno ID")
			}

			listEst, err := estRepo.ListByPagamentoID(ctx, pagamento.ID, 10, 0)
			if err != nil {
				t.Fatalf("ListByPagamentoID: %v", err)
			}
			if len(listEst) != 1 {
				t.Fatalf("expected 1 estorno")
			}

			gotEst, err := estRepo.GetByID(ctx, estorno.ID)
			if err != nil {
				t.Fatalf("GetByID estorno: %v", err)
			}
			if gotEst.PagamentoID != pagamento.ID {
				t.Fatalf("pagamento mismatch")
			}

			// Cobertura de Update e Deletes simples.
			pagamento.Metodo = tcCase.metodoNovo
			if err := payRepo.Update(ctx, &pagamento); err != nil {
				t.Fatalf("Update pagamento: %v", err)
			}

			if err := estRepo.Delete(ctx, estorno.ID); err != nil {
				t.Fatalf("Delete estorno: %v", err)
			}
			_, err = estRepo.GetByID(ctx, estorno.ID)
			if !errors.Is(err, repository.ErrNotFound) {
				t.Fatalf("expected ErrNotFound for estorno, got %v", err)
			}

			if err := payRepo.Delete(ctx, pagamento.ID); err != nil {
				t.Fatalf("Delete pagamento: %v", err)
			}
			_, err = payRepo.GetByID(ctx, pagamento.ID)
			if !errors.Is(err, repository.ErrNotFound) {
				t.Fatalf("expected ErrNotFound for pagamento, got %v", err)
			}
		})
	}
}
