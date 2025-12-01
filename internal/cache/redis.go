package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/GabrielFerrarez19/gofinance-api/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type RedisClient struct {
	client *redis.Client
}

// NewRedisConnection cria uma nova conexão com Redis
func NewRedisConnection(cfg *config.Config) (*RedisClient, error) {
	addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     "", // Sem senha por padrão
		DB:           0,  // Usar DB padrão
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Testar conexão
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info().Msg("Connected to Redis successfully")

	return &RedisClient{
		client: client,
	}, nil
}

// Close fecha a conexão com Redis
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// GetClient retorna o cliente Redis (para uso direto se necessário)
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// Health verifica se o Redis está funcionando
func (r *RedisClient) Health(ctx context.Context) error {
	_, err := r.client.Ping(ctx).Result()
	return err
}
