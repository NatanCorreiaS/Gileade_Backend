# Gileade Backend

API REST para o sistema de venda de ingressos **Gileade Connect**, integrada com Mercado Pago.

## Tecnologias

- **Golang** 1.26+
- **Gin** — framework HTTP
- **Gorm** — ORM com Postgres
- **JWT** — autenticacao stateless
- **Bcrypt** — hash de senhas

## Variaveis de Ambiente

| Variavel | Obrigatoria | Padrao | Descricao |
|---|---|---|---|
| `DB_HOST` | Sim | — | Host do Postgres |
| `DB_PORT` | Sim | — | Porta do Postgres |
| `DB_USER` | Sim | — | Usuario do Postgres |
| `DB_PASSWORD` | Sim | — | Senha do Postgres |
| `DB_NAME` | Sim | — | Nome do banco |
| `DB_SSLMODE` | Nao | `disable` | Modo SSL |
| `DB_TIMEZONE` | Nao | `UTC` | Timezone do banco |
| `APP_PORT` | Nao | `8080` | Porta do servidor HTTP |
| `JWT_SECRET` | Nao | aleatorio | Chave de assinatura dos tokens JWT |
| `JWT_TTL_HOURS` | Nao | `24` | Tempo de vida do token em horas |
| `MERCADO_PAGO_ACCESS_TOKEN_TEST` | Sim | — | Access token do Mercado Pago |
| `MERCADO_PAGO_NOTIFICATION_URL` | Sim | — | URL de webhook do Mercado Pago |
| `AUDIT_LOG_PATH` | Nao | `logs/audit.log` | Caminho do arquivo de auditoria |

## Executar

```bash
# Subir o banco
docker compose up -d

# Executar a API
cp .env.example .env  # preencher as variaveis
go run .
```

## Endpoints

Prefixo base: `/api/v1`

---

### Autenticacao

#### `POST /api/v1/auth/login`

Autentica um usuario por CPF e senha, retornando token JWT e dados do usuario.

**Request:**
```json
{
  "cpf": "12345678900",
  "senha": "minha-senha"
}
```

**Response** `200`:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "usuario": {
    "id": 1,
    "nome": "Joao Silva",
    "tipo_usuario": "Usuario",
    "cpf": "12345678900",
    "idade": 29,
    "celular": "+55 11 99999-0001",
    "igreja": "Igreja Central",
    "papel_igreja": "Membro",
    "estado_civil": "Solteiro(a)",
    "email": "joao@exemplo.com",
    "sexo": "Masculino",
    "cidade": "Sao Paulo",
    "estado_uf": "SP",
    "escolaridade": "Ensino Superior Completo"
  }
}
```

**Erros:**
| Status | Mensagem |
|---|---|
| `400` | `cpf e senha sao obrigatorios` |
| `401` | `cpf ou senha invalidos` |

---

#### `POST /api/v1/auth/logout`

Invalida o token JWT, impedindo seu reuso ate a expiracao.

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response** `200`:
```json
{
  "mensagem": "logout realizado"
}
```

**Erros:**
| Status | Mensagem |
|---|---|
| `400` | `token de autorizacao ausente` |

---

### Pessoas

#### `POST /api/v1/pessoas`

Cadastra uma pessoa no sistema. A senha e automaticamente hasheada com bcrypt.

**Request:**
```json
{
  "nome": "Joao Silva",
  "tipo_usuario": "Usuario",
  "senha": "minha-senha",
  "cpf": "12345678900",
  "idade": 29,
  "celular": "+55 11 99999-0001",
  "igreja": "Igreja Central",
  "papel_igreja": "Membro",
  "estado_civil": "Solteiro(a)",
  "email": "joao@exemplo.com",
  "sexo": "Masculino",
  "cidade": "Sao Paulo",
  "estado_uf": "SP",
  "escolaridade": "Ensino Superior Completo"
}
```

#### `GET /api/v1/pessoas?limit=50&offset=0`

Lista pessoas com paginacao.

#### `GET /api/v1/pessoas/:id`

Busca uma pessoa pelo ID.

#### `PUT /api/v1/pessoas/:id`

Atualiza dados de uma pessoa. Campos enviados como `null` sao ignorados. Se enviar `senha`, ela sera hasheada automaticamente.

#### `DELETE /api/v1/pessoas/:id`

Remove uma pessoa pelo ID.

---

### Tickets

#### `POST /api/v1/tickets`

Cria um tipo de ingresso.

**Request:**
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

**Tipos validos:** `Individual`, `Duo`, `Caravana`

#### `GET /api/v1/tickets?limit=50&offset=0`

Lista tickets com paginacao.

#### `GET /api/v1/tickets/:id`

Busca um ticket pelo ID.

#### `PUT /api/v1/tickets/:id`

Atualiza um ticket.

#### `DELETE /api/v1/tickets/:id`

Remove um ticket.

---

### Tickets Compra

#### `POST /api/v1/tickets-compra`

Cria um vinculo de compra de ticket.

**Request:**
```json
{
  "usuario_id": 1,
  "ticket_id": 1,
  "quantidade": 1,
  "status": "Pendente"
}
```

#### `GET /api/v1/tickets-compra/:id`

Busca uma compra pelo ID.

#### `GET /api/v1/usuarios/:id/tickets-compra?limit=50&offset=0`

Lista compras de um usuario.

#### `PATCH /api/v1/tickets-compra/:id/status`

Atualiza o status de uma compra.

**Request:**
```json
{
  "status": "Pago"
}
```

**Status validos:** `Pendente`, `Pago`, `Cancelado`, `Reembolsado`

#### `DELETE /api/v1/tickets-compra/:id`

Remove uma compra.

---

### Pagamentos

#### `POST /api/v1/pagamentos/checkout`

Cria um checkout no Mercado Pago e persiste o ticket como pendente.

**Request:**
```json
{
  "usuario_id": 1,
  "ticket_id": 1,
  "quantidade": 1,
  "beneficiados": [
    {
      "nome": "Beneficiado 1",
      "cpf": "12345678909",
      "idade": 29,
      "celular": "+55 11 99999-0001",
      "igreja": "Igreja Central",
      "papel_igreja": "Membro",
      "estado_civil": "Solteiro(a)",
      "email": "beneficiado1@exemplo.com",
      "sexo": "Masculino",
      "cidade": "Sao Paulo",
      "estado_uf": "SP",
      "escolaridade": "Ensino Superior Completo"
    }
  ],
  "success_url": "https://seu-site.com/sucesso",
  "failure_url": "https://seu-site.com/erro",
  "pending_url": "https://seu-site.com/pendente"
}
```

A quantidade de beneficiados deve corresponder ao tipo do ticket:
- **Individual:** 1 beneficiado por unidade
- **Duo:** 2 beneficiados por unidade
- **Caravana:** 10 beneficiados por unidade

**Response:**
```json
{
  "preference_id": "123456789-abc...",
  "init_point": "https://www.mercadopago.com.br/...",
  "sandbox_init_point": "https://sandbox.mercadopago.com.br/...",
  "ticket_compra_id": 1
}
```

#### `POST /api/v1/pagamentos/webhook`

Recebe notificacoes de pagamento do Mercado Pago. Processa automaticamente pagamentos aprovados, atualizando o status do ticket para `Pago`.

#### `GET /api/v1/pagamentos?usuario_id=1&status=Pago&limit=50&offset=0`

Lista pagamentos com filtros opcionais.

**Parametros:**
| Parametro | Descricao |
|---|---|
| `usuario_id` | Obrigatorio. IDs separados por virgula |
| `status` | `Pendente`, `Pago`, `Cancelado`, `Reembolsado` |
| `data_inicio` | Data ISO 8601 |
| `data_fim` | Data ISO 8601 |
| `limit` | Padrao 50 |
| `offset` | Padrao 0 |

---

### Estornos

#### `POST /api/v1/pagamentos/:id/estornos`

Cria um estorno (reembolso) via Mercado Pago e atualiza o ticket para `Reembolsado`.

**Request:**
```json
{
  "motivo": "cancelamento",
  "valor": "120.00"
}
```

O campo `valor` e opcional — se omitido, faz estorno total.

---

## Integracao com Flutter

### Fluxo de autenticacao

1. O app envia `POST /api/v1/auth/login` com CPF e senha
2. Em caso de sucesso, armazena o `token` (ex: com `flutter_secure_storage`)
3. As chamadas autenticadas devem incluir o header `Authorization: Bearer <token>`
4. No logout, envia `POST /api/v1/auth/logout` com o header `Authorization`

### Exemplo Dart (http)

```dart
// Login
final response = await http.post(
  Uri.parse('$baseUrl/api/v1/auth/login'),
  headers: {'Content-Type': 'application/json'},
  body: jsonEncode({'cpf': cpf, 'senha': senha}),
);

if (response.statusCode == 200) {
  final data = jsonDecode(response.body);
  final token = data['token'];
  // armazenar token
}

// Logout
await http.post(
  Uri.parse('$baseUrl/api/v1/auth/logout'),
  headers: {'Authorization': 'Bearer $token'},
);
// remover token do armazenamento local
```
