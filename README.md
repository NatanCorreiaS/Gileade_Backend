## Endpoints

Base URL (dev): `http://localhost:8080`

### Criar Checkout Pro

**POST** `/api/v1/checkout/pro`

Cria uma preferencia no Mercado Pago (Checkout Pro) e registra um `TicketUsuario` com status `Pendente`.

**Request JSON**
```json
{
	"usuario_id": 1,
	"ticket_id": 2,
	"success_url": "https://seu-site.com/checkout/sucesso",
	"failure_url": "https://seu-site.com/checkout/erro",
	"pending_url": "https://seu-site.com/checkout/pendente",
	"notification_url": "https://seu-dominio.com/api/v1/mercadopago/webhook"
}
```

**Campos**
- `usuario_id` (obrigatorio): ID da pessoa compradora.
- `ticket_id` (obrigatorio): ID do ticket a ser comprado.
- `success_url` (opcional): URL de retorno quando aprovado.
- `failure_url` (opcional): URL de retorno quando falhar.
- `pending_url` (opcional): URL de retorno quando pendente.
- `notification_url` (opcional): URL do webhook; se vazio, usa `MERCADO_PAGO_NOTIFICATION_URL` do `.env`.

**Response 200**
```json
{
	"preference_id": "1234567890",
	"init_point": "https://www.mercadopago.com.br/checkout/v1/redirect?pref_id=...",
	"sandbox_init_point": "https://sandbox.mercadopago.com.br/checkout/v1/redirect?pref_id=...",
	"ticket_usuario_id": 15
}
```

**Erros comuns**
- `400`: payload invalido ou `notification_url` ausente.
- `404`: usuario ou ticket nao encontrados.
- `502`: falha ao criar preferencia no Mercado Pago.

---

### Webhook Mercado Pago

**POST** `/api/v1/mercadopago/webhook`

Recebe notificacoes do Mercado Pago e registra o pagamento aprovado, marcando o `TicketUsuario` como `Pago` de forma atomica.

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

### Tickets por usuario

**POST** `/api/v1/tickets-usuario`

Cria um vinculo entre usuario e ticket (status default: `Pendente`).

**Request JSON**
```json
{
	"usuario_id": 1,
	"ticket_id": 10,
	"status": "Pendente"
}
```

**GET** `/api/v1/tickets-usuario/:id`

Busca vinculo por ID.

**GET** `/api/v1/usuarios/:id/tickets?limit=50&offset=0`

Lista tickets do usuario.

**PATCH** `/api/v1/tickets-usuario/:id/status`

Atualiza apenas o status.

**Request JSON**
```json
{
	"status": "Pago"
}
```

**DELETE** `/api/v1/tickets-usuario/:id`

Remove o vinculo.

---

### Pagamentos e estornos

O pagamento e registrado automaticamente quando o webhook do Mercado Pago confirma `approved`.
Nao existe endpoint publico para criacao manual de pagamento.

Estornos sao registrados no dominio com transacao e devem manter consistencia com o ticket do usuario.
Se for expor um endpoint de estorno, use transacao e registre auditoria.

---

## Auditoria

Os eventos relevantes de usuarios, tickets e pagamentos sao registrados no arquivo definido por `AUDIT_LOG_PATH`
(padrao: `logs/audit.log`).

Regras de seguranca na auditoria:
- CPF e mascarado.
- Tokens sao registrados apenas com os 2 ultimos caracteres.
- Senhas nunca sao registradas.
