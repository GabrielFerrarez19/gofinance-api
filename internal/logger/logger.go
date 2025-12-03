package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger inicializa o logger da aplicação
// Configura o formato de saída baseado no ambiente:
// - development: saída colorida e formatada no console (mais legível para desenvolvimento)
// - production: saída em JSON estruturado (melhor para sistemas de log agregados)
func InitLogger(env string) {
	// Define o formato de timestamp como RFC3339 (padrão ISO 8601)
	zerolog.TimeFieldFormat = time.RFC3339

	if env == "development" {
		// Modo desenvolvimento: saída colorida e formatada no console
		// TimeFormat define como o timestamp será exibido (HH:MM:SS)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	} else {
		// Modo produção: saída em JSON estruturado com timestamp
		// Facilita integração com ferramentas de log como ELK, Datadog, etc.
		log.Logger = log.With().Timestamp().Logger()
	}
}
