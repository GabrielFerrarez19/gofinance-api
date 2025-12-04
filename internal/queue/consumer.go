package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

// MessageHandler é uma função que processa uma mensagem recebida
// Retorna true se a mensagem foi processada com sucesso (ack)
// Retorna false se houve erro e a mensagem deve ser rejeitado/reenviada
type MessageHandler func(ctx context.Context, body []byte, headers amqp.Table) (bool, error)

// Consumer consome mensagens de uma fila no RabbitMQ
type Consumer struct {
	client *RabbitMQClient
}

// NewConsumer cria uma nova instância do consumer
func NewConsumer(client *RabbitMQClient) *Consumer {
	return &Consumer{
		client: client,
	}
}

// Consume inicia o consumo de mensagens de uma fila
// Processa mensagens de forma assuíncrona usando o handler fornecido
func (c *Consumer) Consume(ctx context.Context, queueName string, handler MessageHandler) error {
	// Declarar fila (garante que existe)
	_, err := c.client.DeclareQueue(queueName, true, false, false, false)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Configurar QoS (Quality of Service)
	// PrefetchCount: quantas mensagens não confirmadas o consumer pode receber
	// PrefetchSize: tamanho máximo (0 = ilimitado)
	// Global: aplicar a todas as conexões ou apenas esta
	err = c.client.GetChannel().Qos(
		1,     // PrefetchCount: processar uma mensagem por voz
		0,     // PrefetchSize: ilimitado
		false, // Global: apenas este consumer
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Iniciar consumo do mensagens
	msgs, err := c.client.GetChannel().Consume(
		queueName, // Nome da fila
		"",        // Consumer tag (vazio = gerado automaticamente)
		false,     // Auto-ack (false = ack manual após processamento)
		false,     // Exclusive (false = múltiplos consumers podem consumir)
		false,     // No-local (false = receber mensagens desta conexão)
		false,     // No-wait (false = aguardar confirmação)
		nil,       // Arguments
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Info().
		Str("queue", queueName).
		Msg("Started consuming messages")

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Consumer context cancelled, stopping...")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Warn().Msg("MAssage channel closed")
					return
				}

				// Processar mensagem
				c.processMessage(ctx, msg, handler)
			}
		}
	}()

	return nil
}

// processMessage processa uma mensagem individual
// Chama o handler e faz ack/nack baseado no resultado
func (c *Consumer) processMessage(ctx context.Context, msg amqp.Delivery, handler MessageHandler) {
	// Criar contexto com timeout para processamento
	processCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Chamar handler para processar a mensagem
	success, err := handler(processCtx, msg.Body, msg.Headers)
	if err != nil {
		log.Error().
			Err(err).
			Str("queue", msg.RoutingKey).
			Msg("Error processing message")

		// Rejeitar mensagem e não reenviar (requeue = false)
		// Em produção, você pode querer implementar dead letter queue
		msg.Nack(false, false)
		return
	}

	if success {
		// Confirmar prossamento bem-sucedido (ack)
		// Remove a mensagem da fila
		msg.Ack(false)
		log.Debug().
			Str("queue", msg.RoutingKey).
			Msg("Message precessed successfully")
	} else {
		// Rejeitar mensagem mas reenviar para a fila (requeue = true)
		// Ùtil para error temporários
		msg.Nack(false, true)
		log.Warn().
			Str("queue", msg.RoutingKey).
			Msg("Message processing failed, requeuing")
	}
}

// UnmarshalMessage deserializa o corpo de uma mensagem JSON para um struct
// Helper útil para handlers
func UnmarshalMessage(body []byte, dest any) error {
	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return nil
}
