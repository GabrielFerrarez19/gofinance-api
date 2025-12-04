#!/bin/bash

# Exemplo rápido de requisição para gerar relatório
# Substitua SEU_TOKEN_AQUI pelo seu token JWT

TOKEN="SEU_TOKEN_AQUI"

curl -X POST http://localhost:8080/api/v1/reports \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "type": "monthly",
    "title": "Relatório Mensal - Janeiro 2024",
    "description": "Análise financeira do mês de janeiro",
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-01-31T23:59:59Z"
  }'


