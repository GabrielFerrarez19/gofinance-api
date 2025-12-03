package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config armazena todas as configurações da aplicação
// As configurações são carregadas de variáveis de ambiente ou valores padrão
type Config struct {
	AppEnv      string // Ambiente da aplicação (development, production, etc)
	AppPort     string // Porta onde o servidor HTTP irá escutar
	DBHost      string // Host do banco de dados PostgreSQL
	DBPort      string // Porta do banco de dados PostgreSQL
	DBUser      string // Usuário do banco de dados
	DBPassword  string // Senha do banco de dados
	DBName      string // Nome do banco de dados
	RedisHost   string // Host do Redis
	RedisPort   string // Porta do Redis
	RabbitMQURL string // URL de conexão do RabbitMQ
	JWTSecret   string // Chave secreta para assinar tokens JWT
}

// LoadConfig carrega as configurações da aplicação
// Primeiro tenta carregar do arquivo .env (se existir), depois lê variáveis de ambiente
// Se uma variável não estiver definida, usa o valor padrão fornecido
func LoadConfig() (*Config, error) {
	// Carrega o arquivo .env (opcional - não mostra erro se não existir)
	// O underscore ignora o erro caso o arquivo não exista
	_ = godotenv.Load(".env")

	return &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		AppPort:     getEnv("APP_PORT", "8080"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "postgres"),
		DBName:      getEnv("DB_NAME", "gofinance"),
		RedisHost:   getEnv("REDIS_HOST", "localhost"),
		RedisPort:   getEnv("REDIS_PORT", "6379"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		JWTSecret:   getEnv("JWT_SECRET", "supersecretkey"),
	}, nil
}

// getEnv obtém o valor de uma variável de ambiente
// Se a variável não existir ou estiver vazia, retorna o valor padrão
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
