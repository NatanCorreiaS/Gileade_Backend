package model

// Enums do domínio, conforme fluxograma "Gileade Connect - Fluxograma de dados".
//
// Observação: optamos por armazenar esses valores como strings no banco
// para manter portabilidade e simplicidade no Gorm.

type Sexo string

const (
	SexoMasculino Sexo = "Masculino"
	SexoFeminino  Sexo = "Feminino"
)

type TipoUsuario string

const (
	TipoUsuarioAdmin   TipoUsuario = "Admin"
	TipoUsuarioUsuario TipoUsuario = "Usuario"
)

type PapelIgreja string

const (
	PapelIgrejaPastor     PapelIgreja = "Pastor"
	PapelIgrejaLider      PapelIgreja = "Líder"
	PapelIgrejaVoluntario PapelIgreja = "Voluntário"
	PapelIgrejaMembro     PapelIgreja = "Membro"
)

type EstadoCivil string

const (
	EstadoCivilSolteiro   EstadoCivil = "Solteiro(a)"
	EstadoCivilCasado     EstadoCivil = "Casado(a)"
	EstadoCivilDivorciado EstadoCivil = "Divorciado(a)"
	EstadoCivilViuvo      EstadoCivil = "Viúvo(a)"
)

type Escolaridade string

const (
	EscolaridadeEnsinoFundamentalIncompleto Escolaridade = "Ensino Fundamental Incompleto"
	EscolaridadeEnsinoFundamentalCompleto   Escolaridade = "Ensino Fundamental Completo"
	EscolaridadeEnsinoMedioIncompleto       Escolaridade = "Ensino Médio Incompleto"
	EscolaridadeEnsinoMedioCompleto         Escolaridade = "Ensino Médio Completo"
	EscolaridadeEnsinoSuperiorIncompleto    Escolaridade = "Ensino Superior Incompleto"
	EscolaridadeEnsinoSuperiorCompleto      Escolaridade = "Ensino Superior Completo"
	EscolaridadePosGraduado                 Escolaridade = "Pós-graduado"
	EscolaridadeMestrado                    Escolaridade = "Mestrado"
	EscolaridadeDoutorado                   Escolaridade = "Doutorado"
)

type EstadoUF string

const (
	EstadoUFAcre             EstadoUF = "AC"
	EstadoUFAlagoas          EstadoUF = "AL"
	EstadoUFAmapa            EstadoUF = "AP"
	EstadoUFAmazonas         EstadoUF = "AM"
	EstadoUFBahia            EstadoUF = "BA"
	EstadoUFCeara            EstadoUF = "CE"
	EstadoUFDistritoFederal  EstadoUF = "DF"
	EstadoUFEspiritoSanto    EstadoUF = "ES"
	EstadoUFGoias            EstadoUF = "GO"
	EstadoUFMaranhao         EstadoUF = "MA"
	EstadoUFMatoGrosso       EstadoUF = "MT"
	EstadoUFMatoGrossoDoSul  EstadoUF = "MS"
	EstadoUFMinasGerais      EstadoUF = "MG"
	EstadoUFPara             EstadoUF = "PA"
	EstadoUFParaiba          EstadoUF = "PB"
	EstadoUFParana           EstadoUF = "PR"
	EstadoUFPernambuco       EstadoUF = "PE"
	EstadoUFPiaui            EstadoUF = "PI"
	EstadoUFRioDeJaneiro     EstadoUF = "RJ"
	EstadoUFRioGrandeDoNorte EstadoUF = "RN"
	EstadoUFRioGrandeDoSul   EstadoUF = "RS"
	EstadoUFRondonia         EstadoUF = "RO"
	EstadoUFRoraima          EstadoUF = "RR"
	EstadoUFSantaCatarina    EstadoUF = "SC"
	EstadoUFSaoPaulo         EstadoUF = "SP"
	EstadoUFSeripe           EstadoUF = "SE"
	EstadoUFTocantins        EstadoUF = "TO"
)

type TicketsStatus string

const (
	TicketsStatusPendente    TicketsStatus = "Pendente"
	TicketsStatusPago        TicketsStatus = "Pago"
	TicketsStatusCancelado   TicketsStatus = "Cancelado"
	TicketsStatusReembolsado TicketsStatus = "Reembolsado"
)

type MetodoPagamento string

const (
	MetodoPagamentoPix           MetodoPagamento = "Pix"
	MetodoPagamentoCartaoCredito MetodoPagamento = "Cartão de Crédito"
	MetodoPagamentoBoleto        MetodoPagamento = "Boleto"
)

type TipoTicket string

const (
	TipoTicketIndividual TipoTicket = "Individual"
	TipoTicketDuo        TipoTicket = "Duo"
	TipoTicketCaravana   TipoTicket = "Caravana"
)
