## Endpoints

Base URL (dev): `http://localhost:8080`
**ENDPOINT PUBLICO PARA TESTES**: `https://gileadebackend-production.up.railway.app`

### Criar Checkout Pro

**POST** `/api/v1/pagamentos/checkout`

Cria uma preferencia no Mercado Pago (Checkout Pro) e registra um `TicketCompra` com status `Pendente`.

**Request JSON**
```json
{
	"usuario_id": 1,
	"ticket_id": 2,
	"quantidade": 1,
	"beneficiados": [
		{
			"nome": "Beneficiado 1",
			"cpf": "12345678901",
			"idade": 29,
			"celular": "+55 11 99999-0001",
			"igreja": "Igreja Central",
			"papel_igreja": "Membro",
			"estado_civil": "Solteiro(a)",
			"email": "beneficiado1@exemplo.com",
			"sexo": "Feminino",
			"cidade": "Sao Paulo",
			"estado_uf": "SP",
			"escolaridade": "Ensino Superior Completo"
		}
	],
	"success_url": "https://seu-site.com/checkout/sucesso",
	"failure_url": "https://seu-site.com/checkout/erro",
	"pending_url": "https://seu-site.com/checkout/pendente"
}
```

**Campos**
- `usuario_id` (obrigatorio): ID da pessoa compradora.
- `ticket_id` (obrigatorio): ID do ticket a ser comprado.
- `quantidade` (opcional): quantidade de unidades do ticket (default: 1).
- `beneficiados` (obrigatorio): lista de beneficiados com dados completos (Individual=1, Duo=2, Caravana=10 por unidade).
- `success_url` (opcional): URL de retorno quando aprovado.
- `failure_url` (opcional): URL de retorno quando falhar.
- `pending_url` (opcional): URL de retorno quando pendente.
O `notification_url` e sempre lido do `.env` via `MERCADO_PAGO_NOTIFICATION_URL`.

**Reaproveitamento de beneficiados**
Se o CPF do beneficiado ja existir, o backend reutiliza o registro existente e nao cria duplicados.

**Response 200**
```json
{
	"preference_id": "1234567890",
	"init_point": "https://www.mercadopago.com.br/checkout/v1/redirect?pref_id=...",
	"sandbox_init_point": "https://sandbox.mercadopago.com.br/checkout/v1/redirect?pref_id=...",
	"ticket_compra_id": 15
}
```

**Erros comuns**
- `400`: payload invalido.
- `500`: `MERCADO_PAGO_NOTIFICATION_URL` nao configurada.
- `404`: usuario ou ticket nao encontrados.
- `502`: falha ao criar preferencia no Mercado Pago.

---

### Webhook Mercado Pago

**POST** `/api/v1/pagamentos/webhook`

Recebe notificacoes do Mercado Pago e registra o pagamento aprovado, marcando o `TicketCompra` como `Pago` de forma atomica.

O endpoint aceita `data.id` por query string ou no corpo. Exemplo de payload basico:

**Request JSON**
```json
{
	"type": "payment",
	"data": {
		"id": "1234567890"
	}
}
```

**Response 200 (aprovado)**
```json
{
	"status": "ok"
}
```

**Response 200 (nao aprovado)**
```json
{
	"status": "pending"
}
```

**Erros comuns**
- `400`: `payment id` ausente ou invalido.
- `502`: falha ao consultar pagamento no Mercado Pago.
- `500`: falha ao registrar pagamento no banco.

---

### Pessoas (usuarios e admins)

**POST** `/api/v1/pessoas`

Cria uma pessoa (usuario/admin). A senha deve estar **hash**.

**Request JSON**
```json
{
	"nome": "Maria Silva",
	"tipo_usuario": "Usuario",
	"senha": "hash_da_senha",
	"cpf": "12345678901",
	"idade": 29,
	"celular": "+55 11 99999-0001",
	"igreja": "Igreja Central",
	"papel_igreja": "Membro",
	"estado_civil": "Solteiro(a)",
	"email": "maria@exemplo.com",
	"sexo": "Feminino",
	"cidade": "Sao Paulo",
	"estado_uf": "SP",
	"escolaridade": "Ensino Superior Completo"
}
```

**Tipos de usuario**
- `Admin`: acesso administrativo (crie pela mesma rota).
- `Usuario`: comprador/participante.

**Response 201**
```json
{
	"id": 1,
	"nome": "Maria Silva",
	"tipo_usuario": "Usuario",
	"cpf": "12345678901",
	"idade": 29,
	"celular": "+55 11 99999-0001",
	"igreja": "Igreja Central",
	"papel_igreja": "Membro",
	"estado_civil": "Solteiro(a)",
	"email": "maria@exemplo.com",
	"sexo": "Feminino",
	"cidade": "Sao Paulo",
	"estado_uf": "SP",
	"escolaridade": "Ensino Superior Completo"
}
```

**GET** `/api/v1/pessoas?limit=50&offset=0`

Lista pessoas paginadas.

**Response 200**
```json
[
	{
		"id": 1,
		"nome": "Maria Silva",
		"tipo_usuario": "Usuario",
		"cpf": "12345678901",
		"idade": 29,
		"celular": "+55 11 99999-0001",
		"igreja": "Igreja Central",
		"papel_igreja": "Membro",
		"estado_civil": "Solteiro(a)",
		"email": "maria@exemplo.com",
		"sexo": "Feminino",
		"cidade": "Sao Paulo",
		"estado_uf": "SP",
		"escolaridade": "Ensino Superior Completo"
	}
]
```

**GET** `/api/v1/pessoas/:id`

Busca pessoa por ID.

**PUT** `/api/v1/pessoas/:id`

Atualiza campos (envie apenas os que quer alterar).

**Request JSON**
```json
{
	"nome": "Maria Silva Souza",
	"email": "maria.souza@exemplo.com"
}
```

**DELETE** `/api/v1/pessoas/:id`

Remove uma pessoa.

---

### Tickets

**POST** `/api/v1/tickets`

Cria um ticket.

**Request JSON**
```json
{
	"tipo": "Individual",
	"nome": "Ingresso Geral",
	"descricao": "Entrada padrao",
	"preco": "120.00",
	"quantidade_disponivel": 100,
	"data_evento": "2026-10-20"
}
```

**Response 201**
```json
{
	"id": 10,
	"tipo": "Individual",
	"nome": "Ingresso Geral",
	"descricao": "Entrada padrao",
	"preco": "120.00",
	"quantidade_disponivel": 100,
	"data_evento": "2026-10-20"
}
```

**GET** `/api/v1/tickets?limit=50&offset=0`

Lista tickets paginados.

**GET** `/api/v1/tickets/:id`

Busca ticket por ID.

**PUT** `/api/v1/tickets/:id`

Atualiza campos (envie apenas os que quer alterar).

**Request JSON**
```json
{
	"descricao": "Entrada padrao (lote 2)",
	"preco": "150.00"
}
```

**DELETE** `/api/v1/tickets/:id`

Remove um ticket.

---

### Tickets por compra

**POST** `/api/v1/tickets-compra`

Cria um vinculo entre usuario e ticket (status default: `Pendente`).

**Request JSON**
```json
{
	"usuario_id": 1,
	"ticket_id": 10,
	"quantidade": 1,
	"status": "Pendente"
}
```

**GET** `/api/v1/tickets-compra/:id`

Busca vinculo por ID.

**GET** `/api/v1/usuarios/:id/tickets-compra?limit=50&offset=0`

Lista tickets do usuario.

**PATCH** `/api/v1/tickets-compra/:id/status`

Atualiza apenas o status.

**Request JSON**
```json
{
	"status": "Pago"
}
```

**DELETE** `/api/v1/tickets-compra/:id`

Remove o vinculo.

---

### Pagamentos e estornos

O pagamento e registrado automaticamente quando o webhook do Mercado Pago confirma `approved`.
Nao existe endpoint publico para criacao manual de pagamento.

Estornos sao registrados no dominio com transacao e devem manter consistencia com o ticket da compra.
Se for expor um endpoint de estorno, use transacao e registre auditoria.

---

### Consulta de pagamentos

**GET** `/api/v1/pagamentos?usuario_id=1&status=Pago&data_inicio=2026-01-01&data_fim=2026-12-31&limit=50&offset=0`

Lista pagamentos por usuario com filtros opcionais.

**Query params**
- `usuario_id` (obrigatorio): um ou mais IDs (separados por virgula).
- `status` (opcional): `Pendente`, `Pago`, `Cancelado`, `Reembolsado`.
- `data_inicio` (opcional): data no formato `YYYY-MM-DD`.
- `data_fim` (opcional): data no formato `YYYY-MM-DD`.
- `limit` (opcional): default 50.
- `offset` (opcional): default 0.

---

## Teste rapido de compra (Checkout Pro)

Passo a passo com apenas os endpoints necessarios para criar usuario e ticket, gerar o checkout e finalizar a compra.

### 1) Criar usuario

**POST** `/api/v1/pessoas`

**Request JSON**
```json
{
	"nome": "Teste Compra",
	"tipo_usuario": "Usuario",
	"senha": "hash_da_senha",
	"cpf": "40402519890",
	"idade": 30,
	"celular": "+55 11 99999-0001",
	"igreja": "Igreja Teste",
	"papel_igreja": "Membro",
	"estado_civil": "Solteiro(a)",
	"email": "teste.compra@exemplo.com",
	"sexo": "Masculino",
	"cidade": "Sao Paulo",
	"estado_uf": "SP",
	"escolaridade": "Ensino Superior Completo"
}
```

Guarde o `id` retornado para usar como `usuario_id`.

### 2) Criar ticket

**POST** `/api/v1/tickets`

**Request JSON**
```json
{
	"tipo": "Individual",
	"nome": "Ingresso Teste",
	"descricao": "Entrada para teste rapido",
	"preco": "10.00",
	"quantidade_disponivel": 10,
	"data_evento": "2026-10-20"
}
```

Guarde o `id` retornado para usar como `ticket_id`.

### 3) Criar Checkout Pro

**POST** `/api/v1/pagamentos/checkout`

**Request JSON**
```json
{
	"usuario_id": 1,
	"ticket_id": 2,
	"quantidade": 1,
	"beneficiados": [
		{
			"nome": "Beneficiado 1",
			"cpf": "12345678901",
			"idade": 29,
			"celular": "+55 11 99999-0001",
			"igreja": "Igreja Central",
			"papel_igreja": "Membro",
			"estado_civil": "Solteiro(a)",
			"email": "beneficiado1@exemplo.com",
			"sexo": "Feminino",
			"cidade": "Sao Paulo",
			"estado_uf": "SP",
			"escolaridade": "Ensino Superior Completo"
		}
	],
	"success_url": "https://seu-site.com/checkout/sucesso",
	"failure_url": "https://seu-site.com/checkout/erro",
	"pending_url": "https://seu-site.com/checkout/pendente"
}
```

Use os IDs retornados nos passos 1 e 2. Copie o `init_point` da resposta.

### 4) Abrir o Checkout Pro

- Abra o `init_point` em uma aba anonima.
- Faca login com:
  - usuario: `TESTUSER4040251989076800435`
  - senha: `u1y8ImOFiV`
  - codigo de verificacao: `672106` (caso seja solicitado algum codigo de email)

Se apos logar nao for redirecionado para a compra, abra novamente o link do `init_point`.

### 5) Finalizar o pagamento

- Use qualquer opcao de cartao de credito ou saldo.
- Se usar cartao e for solicitado:
  - senha de seguranca: `123`
  - validade: `11/30`

---

## Auditoria

Os eventos relevantes de usuarios, tickets e pagamentos sao registrados no arquivo definido por `AUDIT_LOG_PATH`
(padrao: `logs/audit.log`).

Regras de seguranca na auditoria:
- CPF e mascarado.
- Tokens sao registrados apenas com os 2 ultimos caracteres.
- Senhas nunca sao registradas.
