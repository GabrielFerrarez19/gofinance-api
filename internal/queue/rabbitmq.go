package queue

import (
	"fmt"
	"time"

	"github.com/GabrielFerrarez19/gofinance-api/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

// RabiitMQClient encapsula a conexão e canal do RabbitMQ
type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *config.Config
}

// NewRabbitMQConnection cria uma nova conexão com RabbitMQ
// Estabelece conexão e canal para publicar/consumir mensagens
// Implementa retry logic para aguardar o RabbitMQ estar pronto
func NewRabbitMQConnection(cfg *config.Config) (*RabbitMQClient, error) {
	var conn *amqp.Connection
	var err error

	// Tentar conectar com retry (até 30 segundos)
	maxRetries := 30
	retryInterval := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(cfg.RabbitMQURL)
		if err == nil {
			break
		}

		log.Warn().
			Err(err).
			Int("attempt", i+1).
			Int("max_retries", maxRetries).
			Msg("Failed to connect to RabbitMQ, retrying...")

		if i < maxRetries-1 {
			time.Sleep(retryInterval)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", maxRetries, err)
	}

	// Criar canal de comunicação
	// O canal é usado para todas as operações (publish, consume, declare queues)
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	log.Info().Msg("Connected to RabbitMQ successfully")

	return &RabbitMQClient{
		conn:    conn,
		channel: channel,
		config:  cfg,
	}, nil
}

// GetChannel retorna o canal do RabbitMQ
// Usado por publishers e consumers para operações
func (r *RabbitMQClient) GetChannel() *amqp.Channel {
	return r.channel
}

// GetConnection retorna a conexão do RabbitMQ
func (r *RabbitMQClient) GetConnection() *amqp.Connection {
	return r.conn
}

// Close decha a conexão e canal do RabbitMQ
// Deve ser chamado ao finalizar a aplicação
func (r *RabbitMQClient) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			return err
		}
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// Health verifica se a conexão com RabbitMQ está saudável
func (r *RabbitMQClient) Health() error {
	if r.conn == nil || r.conn.IsClosed() {
		return fmt.Errorf("RabbitMQ connection is closed")
	}
	return nil
}

// Reconnect tenta reconectar ao RabbitMQ em caso de falha
// Útil para lidar com desconexões temporárias
func (r *RabbitMQClient) Reconnect() error {
	// Fechar conexão antiga se existir
	if r.conn != nil && !r.conn.IsClosed() {
		r.conn.Close()
	}

	// Tentar reconectar
	conn, err := amqp.Dial(r.config.RabbitMQURL)
	if err != nil {
		return fmt.Errorf("failed to reconnect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to reopen channel: %w", err)
	}

	r.conn = conn
	r.channel = channel

	log.Info().Msg("Reconnected to RabbitMQ successfully")
	return nil
}

// DeclareQueue declara uma fila no RabbitMQ
// Se a fila não existir, será criada, Se existir, será reutilizada
func (r *RabbitMQClient) DeclareQueue(name string, durable, autoDelete, exclusive, noWait bool) (amqp.Queue, error) {
	return r.channel.QueueDeclare(
		name,       // Nome da fila
		durable,    // Persistente (sobrevive a reinicializações do servidor)
		autoDelete, // Auto-deletar quando não houver consumidores
		exclusive,  // Exclusiva para esta conexão
		noWait,     // Não aguardar confirmação
		nil,
	)
}

// DeclareExchange declara um exchange no RabbitMQ
// Exchanges roteiam mensagens para filas baseado em regras
func (r *RabbitMQClient) DeclareExchange(name, kind string, durable, autoDelete, internal, noWait bool) error {
	return r.channel.ExchangeDeclare(
		name,       // Nome do exchange
		kind,       // Tipo (direct, topic, fanout, headers)
		durable,    // Persistente
		autoDelete, // Auto-deletar quando não usado
		internal,   // Apenas para uso internos (não aceita publish externo)
		noWait,     // Não aguardar confirmação
		nil,
	)
}
