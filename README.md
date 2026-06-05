# Gileade Backend

API REST para o sistema de venda de ingressos **Gileade Connect**, integrada com Mercado Pago.

## NOTA PARA A PROFESSORA!

Professora, eu tentei de diversas maneiras fazer o estorno e o cancelamento funcionarem com as credenciais de teste e não obtive êxito algum, então decidi por remover os endpoints de `cancelamento` e `estorno`, infelizmente eu não tive como prever isso até chegar nessa parte, a documentação do mercado pago sobre esses dois é realmente confusa, espero que compreenda esse problema.

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
| `ADMIN_PASSWORD` | Nao | — | Senha do usuario administrador (`admin`) |
| `BACKUP_PASSWORD` | Nao | — | Senha do usuario administrador reserva (`backup`) |

## Usuarios Administradores

Na primeira execucao do servidor, dois usuarios administradores sao criados automaticamente se ainda nao existirem:

| Usuario | CPF | Senha (env) |
|---|---|---|
| `Administrador` | `00000000191` | `ADMIN_PASSWORD` |
| `Backup Admin` | `00000000291` | `BACKUP_PASSWORD` |

Se os usuarios ja existirem no banco, a criacao e ignorada. As senhas devem ser definidas no `.env`; caso contrario, os respectivos admins nao serao criados.

## Executar

```bash
# Subir o banco
docker compose up -d

# Executar a API
cp .env.example .env  # preencher as variaveis
go run .
```

## Autenticacao e Autorizacao

Toda rota de escrita (POST/PUT/PATCH/DELETE) requer autenticacao via token JWT no header `Authorization: Bearer <token>`, exceto:

- `POST /api/v1/auth/login` — login
- `POST /api/v1/pessoas` — cadastro publico de usuario
- `POST /api/v1/pagamentos/webhook` — webhook do Mercado Pago

Rotas de leitura (GET) sao publicas.

### Regras de acesso

| Acao | Usuario comum | Admin |
|---|---|---|
| Criar usuario | Sempre criado como `Usuario` | Sempre criado como `Usuario` |
| Alterar proprio cadastro | Sim (exceto `tipo_usuario`) | Sim |
| Alterar cadastro de outro usuario | Nao | Sim (inclusive `tipo_usuario`) |
| Remover usuario | Apenas a propria conta | Qualquer usuario |
| Criar/alterar/remover tickets | Nao | Sim |
| Criar ticket-compra | Sim (apenas para si) | Sim |
| Alterar status ticket-compra | Nao | Sim |
| Criar checkout | Sim | Sim |

### Cargos (tipo_usuario)

| Valor | Descricao |
|---|---|
| `Usuario` | Usuario comum |
| `Admin` | Administrador com acesso total |

> O campo `tipo_usuario` **nao pode ser definido na criacao** de usuario. O cargo `Admin` so pode ser atribuido por outro administrador atraves da rota `PUT /api/v1/pessoas/:id`.

---

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

#### `POST /api/v1/pessoas` *(publico)*

Cadastra uma pessoa no sistema. O cargo e sempre criado como `Usuario`. A senha e automaticamente hasheada com bcrypt.

**Headers:** Nenhum (rota publica).

**Request:**
```json
{
  "nome": "Joao Silva",
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

**Response** `201`: mesmo formato do login (sem o token).

**Erros:**
| Status | Mensagem |
|---|---|
| `400` | `payload invalido` |
| `409` | `cpf ja cadastrado` |

#### `GET /api/v1/pessoas?limit=50&offset=0` *(publico)*

Lista pessoas com paginacao.

#### `GET /api/v1/pessoas/:id` *(publico)*

Busca uma pessoa pelo ID.

#### `PUT /api/v1/pessoas/:id` *(autenticado)*

Atualiza dados de uma pessoa. Campos enviados como `null` sao ignorados.

**Regras:**
- Usuarios comuns so podem alterar os proprios dados e **nao podem** alterar o campo `tipo_usuario`.
- Administradores podem alterar qualquer usuario, inclusive promover a `Admin` via `tipo_usuario`.

**Headers:**
```
Authorization: Bearer <token>
```

**Request (usuario comum):**
```json
{
  "nome": "Joao Silva Atualizado",
  "email": "novo@exemplo.com"
}
```

**Request (admin promovendo usuario):**
```json
{
  "tipo_usuario": "Admin"
}
```

**Erros:**
| Status | Mensagem |
|---|---|
| `401` | `autenticacao necessaria` |
| `403` | `voce so pode alterar seus proprios dados` |
| `403` | `apenas administradores podem alterar o cargo` |
| `409` | `cpf ja cadastrado` |

#### `DELETE /api/v1/pessoas/:id` *(autenticado)*

Remove uma pessoa pelo ID. Usuarios comuns podem remover apenas a propria conta. Admins podem remover qualquer usuario.

**Headers:**
```
Authorization: Bearer <token>
```

**Erros:**
| Status | Mensagem |
|---|---|
| `401` | `autenticacao necessaria` |
| `403` | `voce so pode remover sua propria conta ou ser um administrador` |
| `404` | `usuario nao encontrado` |

---

### Tickets

#### `POST /api/v1/tickets` *(autenticado)*

Cria um tipo de ingresso.

**Headers:**
```
Authorization: Bearer <token>
```

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

#### `GET /api/v1/tickets?limit=50&offset=0` *(publico)*

Lista tickets com paginacao.

#### `GET /api/v1/tickets/:id` *(publico)*

Busca um ticket pelo ID.

#### `PUT /api/v1/tickets/:id` *(autenticado)*

Atualiza um ticket.

**Headers:**
```
Authorization: Bearer <token>
```

#### `DELETE /api/v1/tickets/:id` *(autenticado)*

Remove um ticket.

**Headers:**
```
Authorization: Bearer <token>
```

---

### Tickets Compra

#### `POST /api/v1/tickets-compra` *(autenticado)*

Cria um vinculo de compra de ticket.

**Headers:**
```
Authorization: Bearer <token>
```

**Request:**
```json
{
  "usuario_id": 1,
  "ticket_id": 1,
  "quantidade": 1,
  "status": "Pendente"
}
```

#### `GET /api/v1/tickets-compra/:id` *(publico)*

Busca uma compra pelo ID.

#### `GET /api/v1/usuarios/:id/tickets-compra?limit=50&offset=0` *(publico)*

Lista compras de um usuario.

#### `PATCH /api/v1/tickets-compra/:id/status` *(autenticado)*

Atualiza o status de uma compra.

**Headers:**
```
Authorization: Bearer <token>
```

**Request:**
```json
{
  "status": "Pago"
}
```

**Status validos:** `Pendente`, `Pago`, `Cancelado`, `Reembolsado`

#### `DELETE /api/v1/tickets-compra/:id` *(autenticado)*

Remove uma compra.

**Headers:**
```
Authorization: Bearer <token>
```

---

### Pagamentos

#### `POST /api/v1/pagamentos/checkout` *(autenticado)*

Cria um checkout no Mercado Pago e persiste o ticket como pendente.

> **Importante sobre estoque:** O ticket **nao e reservado** durante o checkout — ele permanece
> disponivel para outros usuarios ate a confirmacao do pagamento. A disponibilidade e verificada
> no checkout (retorna `409` se insuficiente) e confirmada atomicamente no webhook de pagamento.
> Se o ticket esgotar entre o checkout e o pagamento, a confirmacao sera rejeitada.

**Headers:**
```
Authorization: Bearer <token>
```

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
  "back_urls": {
    "success": "tuapp://success",
    "failure": "tuapp://failure",
    "pending": "tuapp://pending"
  },
  "auto_return": "approved"
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

**Erros:**
| Status | Mensagem |
|---|---|
| `409` | `tickets indisponiveis para a quantidade solicitada` |

#### `POST /api/v1/pagamentos/webhook` *(publico)*

Recebe notificacoes de pagamento do Mercado Pago. Processa automaticamente pagamentos aprovados, atualizando o status do ticket para `Pago`.

#### `GET /api/v1/pagamentos?usuario_id=1&status=Pago&limit=50&offset=0` *(publico)*

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

### Exportacao CSV (Admin)

Endpoints para exportacao de dados em formato CSV. Todos exigem autenticacao de administrador.

> **Requisitos:** `Authorization: Bearer <token>` com cargo `Admin`.

#### `GET /api/v1/admin/export/usuarios` *(admin)*

Exporta todos os usuarios cadastrados em CSV.

**Headers:**
```
Authorization: Bearer <token>
```

**Colunas:** `ID`, `Nome`, `TipoUsuario`, `CPF`, `Idade`, `Celular`, `Igreja`, `PapelIgreja`, `EstadoCivil`, `Email`, `Sexo`, `Cidade`, `EstadoUF`, `Escolaridade`, `DataCriacao`, `DataAtualizacao`

---

#### `GET /api/v1/admin/export/pagamentos` *(admin)*

Exporta todos os pagamentos realizados em CSV, incluindo dados da compra associada.

**Headers:**
```
Authorization: Bearer <token>
```

**Colunas:** `ID`, `IDTransacao`, `Valor`, `Metodo`, `DataPagamento`, `TicketCompraID`, `UsuarioID`, `TicketID`, `Status`

---

#### `GET /api/v1/admin/export/tickets` *(admin)*

Exporta todos os tickets cadastrados em CSV.

**Headers:**
```
Authorization: Bearer <token>
```

**Colunas:** `ID`, `Tipo`, `Nome`, `Descricao`, `Preco`, `QuantidadeDisponivel`, `DataEvento`, `DataCriacao`, `DataAtualizacao`

---

#### `GET /api/v1/admin/export/tickets-compra` *(admin)*

Exporta todas as compras de tickets em CSV, incluindo nome e CPF do comprador.

**Headers:**
```
Authorization: Bearer <token>
```

**Colunas:** `ID`, `UsuarioID`, `UsuarioNome`, `UsuarioCPF`, `Status`, `PreferenceID`, `TicketID`, `TicketNome`, `Quantidade`, `DataCriacao`, `DataAtualizacao`

---

#### `GET /api/v1/admin/export/beneficiados` *(admin)*

Exporta todos os beneficiados cadastrados em CSV.

**Headers:**
```
Authorization: Bearer <token>
```

**Colunas:** `ID`, `Nome`, `CPF`, `Idade`, `Celular`, `Igreja`, `PapelIgreja`, `EstadoCivil`, `Email`, `Sexo`, `Cidade`, `EstadoUF`, `Escolaridade`, `DataCriacao`, `DataAtualizacao`

**Erros comuns a todas as rotas de exportacao:**
| Status | Mensagem |
|---|---|
| `401` | `token de autorizacao ausente` / `token invalido` |
| `403` | `acesso restrito a administradores` |
| `500` | `falha ao consultar ...` |

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
