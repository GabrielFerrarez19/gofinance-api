# Exemplos de Requisições para Gerar Relatórios

## Endpoint

```
POST http://localhost:8080/api/v1/reports
```

## Autenticação

Todas as requisições requerem um token JWT no header:

```
Authorization: Bearer <seu_token_jwt>
```

---

## 1. Relatório Mensal (Monthly)

### cURL

```bash
curl -X POST http://localhost:8080/api/v1/reports \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer SEU_TOKEN_AQUI" \
  -d '{
    "type": "monthly",
    "title": "Relatório Mensal - Janeiro 2024",
    "description": "Análise financeira do mês de janeiro",
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-01-31T23:59:59Z"
  }'
```

### JSON Body

```json
{
  "type": "monthly",
  "title": "Relatório Mensal - Janeiro 2024",
  "description": "Análise financeira do mês de janeiro",
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2024-01-31T23:59:59Z"
}
```

---

## 2. Relatório Anual (Yearly)

### cURL

```bash
curl -X POST http://localhost:8080/api/v1/reports \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer SEU_TOKEN_AQUI" \
  -d '{
    "type": "yearly",
    "title": "Relatório Anual 2024",
    "description": "Visão geral das finanças do ano de 2024",
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-12-31T23:59:59Z"
  }'
```

### JSON Body

```json
{
  "type": "yearly",
  "title": "Relatório Anual 2024",
  "description": "Visão geral das finanças do ano de 2024",
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2024-12-31T23:59:59Z"
}
```

---

## 3. Relatório Personalizado (Custom)

### cURL

```bash
curl -X POST http://localhost:8080/api/v1/reports \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer SEU_TOKEN_AQUI" \
  -d '{
    "type": "custom",
    "title": "Relatório Trimestral Q1 2024",
    "description": "Análise do primeiro trimestre",
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-03-31T23:59:59Z"
  }'
```

### JSON Body

```json
{
  "type": "custom",
  "title": "Relatório Trimestral Q1 2024",
  "description": "Análise do primeiro trimestre",
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2024-03-31T23:59:59Z"
}
```

---

## 4. Relatório Sem Descrição (Opcional)

### cURL

```bash
curl -X POST http://localhost:8080/api/v1/reports \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer SEU_TOKEN_AQUI" \
  -d '{
    "type": "monthly",
    "title": "Relatório Janeiro",
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-01-31T23:59:59Z"
  }'
```

### JSON Body

```json
{
  "type": "monthly",
  "title": "Relatório Janeiro",
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2024-01-31T23:59:59Z"
}
```

---

## Resposta de Sucesso (201 Created)

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "type": "monthly",
  "title": "Relatório Mensal - Janeiro 2024",
  "description": "Análise financeira do mês de janeiro",
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2024-01-31T23:59:59Z",
  "data": null,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Nota:** O campo `data` estará `null` inicialmente. Ele será preenchido pelo job assíncrono do RabbitMQ após o processamento.

---

## Validações

- **type**: Obrigatório. Deve ser um dos valores: `"monthly"`, `"yearly"` ou `"custom"`
- **title**: Obrigatório. Mínimo 2 caracteres, máximo 100 caracteres
- **description**: Opcional. Máximo 500 caracteres
- **start_date**: Obrigatório. Formato ISO 8601 (ex: `2024-01-01T00:00:00Z`)
- **end_date**: Obrigatório. Formato ISO 8601 (ex: `2024-01-31T23:59:59Z`)

---

## Como Obter o Token JWT

Primeiro, faça login para obter o token:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "seu_email@example.com",
    "password": "sua_senha"
  }'
```

A resposta incluirá um `access_token` que deve ser usado no header `Authorization`.

