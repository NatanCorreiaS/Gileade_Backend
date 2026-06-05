package controller

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	model "gileade/gileade_backend/Model"
	"gileade/gileade_backend/audit"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ExportController disponibiliza endpoints de exportacao de dados em CSV para administradores.
type ExportController struct {
	db *gorm.DB
}

// NewExportController instancia o controller de exportacao.
func NewExportController(db *gorm.DB) *ExportController {
	return &ExportController{db: db}
}

// ---- Helpers ----

func writeCSVResponse(ctx *gin.Context, filename string, records [][]string) {
	ctx.Header("Content-Type", "text/csv; charset=utf-8")
	ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	writer := csv.NewWriter(ctx.Writer)
	if err := writer.WriteAll(records); err != nil {
		audit.GetLogger().LogEvent("export_csv_erro", false, map[string]any{
			"arquivo": filename,
		}, err)
	}
}

func fmtTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// ---- ExportPessoas exporta todos os usuarios em CSV. ----

func (c *ExportController) ExportPessoas(ctx *gin.Context) {
	audit.GetLogger().LogEvent("export_pessoas_inicio", true, nil, nil)

	var pessoas []model.Pessoa
	if err := c.db.WithContext(ctx).Order("id asc").Find(&pessoas).Error; err != nil {
		audit.GetLogger().LogEvent("export_pessoas_erro", false, nil, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao consultar usuarios"})
		return
	}

	records := [][]string{
		{"ID", "Nome", "TipoUsuario", "CPF", "Idade", "Celular", "Igreja", "PapelIgreja",
			"EstadoCivil", "Email", "Sexo", "Cidade", "EstadoUF", "Escolaridade",
			"DataCriacao", "DataAtualizacao"},
	}
	for _, p := range pessoas {
		records = append(records, []string{
			fmt.Sprintf("%d", p.ID),
			p.Nome,
			string(p.TipoUsuario),
			p.CPF,
			fmt.Sprintf("%d", p.Idade),
			p.Celular,
			p.Igreja,
			string(p.PapelIgreja),
			string(p.EstadoCivil),
			p.Email,
			string(p.Sexo),
			p.Cidade,
			string(p.EstadoUF),
			string(p.Escolaridade),
			fmtTime(p.DataCriacao),
			fmtTime(p.DataAtualizacao),
		})
	}

	audit.GetLogger().LogEvent("export_pessoas_sucesso", true, map[string]any{
		"total": len(pessoas),
	}, nil)
	writeCSVResponse(ctx, "usuarios.csv", records)
}

// ---- ExportPagamentos exporta todos os pagamentos em CSV. ----

func (c *ExportController) ExportPagamentos(ctx *gin.Context) {
	audit.GetLogger().LogEvent("export_pagamentos_inicio", true, nil, nil)

	var pagamentos []model.Pagamento
	if err := c.db.WithContext(ctx).
		Preload("TicketCompra").
		Order("pagamentos.id asc").
		Find(&pagamentos).Error; err != nil {
		audit.GetLogger().LogEvent("export_pagamentos_erro", false, nil, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao consultar pagamentos"})
		return
	}

	records := [][]string{
		{"ID", "IDTransacao", "Valor", "Metodo", "DataPagamento",
			"TicketCompraID", "UsuarioID", "TicketID", "Status"},
	}
	for _, p := range pagamentos {
		records = append(records, []string{
			fmt.Sprintf("%d", p.ID),
			p.IDTransacao,
			p.Valor.StringFixed(2),
			string(p.Metodo),
			fmtTime(p.DataPagamento),
			fmt.Sprintf("%d", p.TicketCompraID),
			fmt.Sprintf("%d", p.TicketCompra.UsuarioID),
			fmt.Sprintf("%d", p.TicketCompra.TicketID),
			string(p.TicketCompra.Status),
		})
	}

	audit.GetLogger().LogEvent("export_pagamentos_sucesso", true, map[string]any{
		"total": len(pagamentos),
	}, nil)
	writeCSVResponse(ctx, "pagamentos.csv", records)
}

// ---- ExportTickets exporta todos os tickets em CSV. ----

func (c *ExportController) ExportTickets(ctx *gin.Context) {
	audit.GetLogger().LogEvent("export_tickets_inicio", true, nil, nil)

	var tickets []model.Ticket
	if err := c.db.WithContext(ctx).Order("id asc").Find(&tickets).Error; err != nil {
		audit.GetLogger().LogEvent("export_tickets_erro", false, nil, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao consultar tickets"})
		return
	}

	records := [][]string{
		{"ID", "Tipo", "Nome", "Descricao", "Preco", "QuantidadeDisponivel",
			"DataEvento", "DataCriacao", "DataAtualizacao"},
	}
	for _, t := range tickets {
		records = append(records, []string{
			fmt.Sprintf("%d", t.ID),
			string(t.Tipo),
			t.Nome,
			t.Descricao,
			t.Preco.StringFixed(2),
			fmt.Sprintf("%d", t.QuantidadeDisponivel),
			t.DataEvento.Format("2006-01-02"),
			fmtTime(t.DataCriacao),
			fmtTime(t.DataAtualizacao),
		})
	}

	audit.GetLogger().LogEvent("export_tickets_sucesso", true, map[string]any{
		"total": len(tickets),
	}, nil)
	writeCSVResponse(ctx, "tickets.csv", records)
}

// ---- ExportTicketsCompra exporta todas as compras de ticket em CSV. ----

func (c *ExportController) ExportTicketsCompra(ctx *gin.Context) {
	audit.GetLogger().LogEvent("export_tickets_compra_inicio", true, nil, nil)

	var compras []model.TicketCompra
	if err := c.db.WithContext(ctx).
		Preload("Usuario").
		Preload("Ticket").
		Order("tickets_compra.id asc").
		Find(&compras).Error; err != nil {
		audit.GetLogger().LogEvent("export_tickets_compra_erro", false, nil, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao consultar compras de tickets"})
		return
	}

	records := [][]string{
		{"ID", "UsuarioID", "UsuarioNome", "UsuarioCPF", "Status", "PreferenceID",
			"TicketID", "TicketNome", "Quantidade", "DataCriacao", "DataAtualizacao"},
	}
	for _, tc := range compras {
		records = append(records, []string{
			fmt.Sprintf("%d", tc.ID),
			fmt.Sprintf("%d", tc.UsuarioID),
			tc.Usuario.Nome,
			tc.Usuario.CPF,
			string(tc.Status),
			tc.PreferenceID,
			fmt.Sprintf("%d", tc.TicketID),
			tc.Ticket.Nome,
			fmt.Sprintf("%d", tc.Quantidade),
			fmtTime(tc.DataCriacao),
			fmtTime(tc.DataAtualizacao),
		})
	}

	audit.GetLogger().LogEvent("export_tickets_compra_sucesso", true, map[string]any{
		"total": len(compras),
	}, nil)
	writeCSVResponse(ctx, "tickets-compra.csv", records)
}

// ---- ExportBeneficiados exporta todos os beneficiados em CSV. ----

func (c *ExportController) ExportBeneficiados(ctx *gin.Context) {
	audit.GetLogger().LogEvent("export_beneficiados_inicio", true, nil, nil)

	var beneficiados []model.Beneficiado
	if err := c.db.WithContext(ctx).Order("id asc").Find(&beneficiados).Error; err != nil {
		audit.GetLogger().LogEvent("export_beneficiados_erro", false, nil, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"erro": "falha ao consultar beneficiados"})
		return
	}

	records := [][]string{
		{"ID", "Nome", "CPF", "Idade", "Celular", "Igreja", "PapelIgreja",
			"EstadoCivil", "Email", "Sexo", "Cidade", "EstadoUF", "Escolaridade",
			"DataCriacao", "DataAtualizacao"},
	}
	for _, b := range beneficiados {
		records = append(records, []string{
			fmt.Sprintf("%d", b.ID),
			b.Nome,
			b.CPF,
			fmt.Sprintf("%d", b.Idade),
			b.Celular,
			b.Igreja,
			string(b.PapelIgreja),
			string(b.EstadoCivil),
			b.Email,
			string(b.Sexo),
			b.Cidade,
			string(b.EstadoUF),
			string(b.Escolaridade),
			fmtTime(b.DataCriacao),
			fmtTime(b.DataAtualizacao),
		})
	}

	audit.GetLogger().LogEvent("export_beneficiados_sucesso", true, map[string]any{
		"total": len(beneficiados),
	}, nil)
	writeCSVResponse(ctx, "beneficiados.csv", records)
}
