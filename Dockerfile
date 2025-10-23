# Stage 1: Build
FROM golang:1.25.2-alpine AS builder  

WORKDIR /app

# Copia dependências e baixa módulos
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código
COPY . .

# Compila o binário
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gofinance ./cmd/api/main.go

# Verifica se o binário foi criado
RUN ls -la gofinance

# Stage 2: Container final
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=builder /app/gofinance .

RUN chmod +x gofinance

EXPOSE 8080

CMD ["./gofinance"]
