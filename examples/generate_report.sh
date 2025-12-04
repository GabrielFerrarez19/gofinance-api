#!/bin/bash

# Script para gerar relatórios na API GoFinance
# Uso: ./generate_report.sh <token_jwt> <tipo> <titulo> <start_date> <end_date> [description]

# Cores para output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Verificar se o token foi fornecido
if [ -z "$1" ]; then
    echo -e "${RED}Erro: Token JWT não fornecido${NC}"
    echo "Uso: $0 <token_jwt> <tipo> <titulo> <start_date> <end_date> [description]"
    echo ""
    echo "Exemplo:"
    echo "  $0 'seu_token_aqui' monthly 'Relatório Janeiro' '2024-01-01T00:00:00Z' '2024-01-31T23:59:59Z' 'Descrição opcional'"
    exit 1
fi

TOKEN=$1
TYPE=${2:-monthly}
TITLE=${3:-"Relatório Mensal"}
START_DATE=${4:-"2024-01-01T00:00:00Z"}
END_DATE=${5:-"2024-01-31T23:59:59Z"}
DESCRIPTION=${6:-""}

API_URL="http://localhost:8080/api/v1/reports"

# Construir JSON body
if [ -z "$DESCRIPTION" ]; then
    JSON_BODY=$(cat <<EOF
{
  "type": "$TYPE",
  "title": "$TITLE",
  "start_date": "$START_DATE",
  "end_date": "$END_DATE"
}
EOF
)
else
    JSON_BODY=$(cat <<EOF
{
  "type": "$TYPE",
  "title": "$TITLE",
  "description": "$DESCRIPTION",
  "start_date": "$START_DATE",
  "end_date": "$END_DATE"
}
EOF
)
fi

echo -e "${YELLOW}Enviando requisição para gerar relatório...${NC}"
echo "URL: $API_URL"
echo "Tipo: $TYPE"
echo "Título: $TITLE"
echo ""

# Fazer a requisição
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "$JSON_BODY")

# Separar body e status code
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

# Verificar status code
if [ "$HTTP_CODE" -eq 201 ]; then
    echo -e "${GREEN}✓ Relatório criado com sucesso!${NC}"
    echo ""
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
else
    echo -e "${RED}✗ Erro ao criar relatório (HTTP $HTTP_CODE)${NC}"
    echo ""
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
    exit 1
fi


