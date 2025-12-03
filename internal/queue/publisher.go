package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

// Publisher publica mensagens em filas do RabbitMQ
type Publisher struct {
	client *RabbitMQClient
}

// NewPublisher cria uma nova instância do publisher
func NewPublisher(client *RabbitMQClient) *Publisher {
	return &Publisher{
		client: client,
	}
}

// Publish envia uma mensagem para uma fila
// Converte o payload para JSON e publica com headers opcionais
func (p *Publisher) Publish(ctx context.Context, queueName string, payload any, headers map[string]any) error {
	// Converter payload para JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Preparar headers da mensagem
	amqpHeaders := make(amqp.Table)
	if headers != nil {
		for k, v := range headers {
			amqpHeaders[k] = v
		}
	}

	// Publicar mensagem
	err = p.client.GetChannel().PublishWithContext(
		ctx,
		"",        // Exchange (vazio = default exchange)
		queueName, // Routing key (nome da fila quando usando default exchange)
		false,     // Mandatory (retornar erro se fila não existir)
		false,     // Immediate (retornar err se não houver consumidor)
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // Mensagem persistente (sobrevive a reinicialização)
			Timestamp:    time.Now(),
			Body:         body,
			Headers:      amqpHeaders,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Debug().Str("queue", queueName).Msg("Message published successfully")

	return nil
}

// PublishToExchange publica uma mensagem para um exchange
// Útil para roteamento mais compelxo(topic, fanout, etc)
func (p *Publisher) PublishToExchange(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	amqpHeaders := make(amqp.Table)
	if headers != nil {
		for k, v := range headers {
			amqpHeaders[k] = v
		}
	}

	err = p.client.GetChannel().PublishWithContext(
		ctx,
		exchange,   // Nome do exchange
		routingKey, // Routing key (usado para rotear mensagem)
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
			Headers:      amqpHeaders,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish to exchange: %w", err)
	}

	log.Debug().
		Str("exchange", exchange).
		Str("routing_key", routingKey).
		Msg("Message published to exchange successfully")

	return nil
}
