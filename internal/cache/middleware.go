package cache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CacheMiddleware struct {
	cache *CacheService
}

func NewCacheMiddleware(cache *CacheService) *CacheMiddleware {
	return &CacheMiddleware{
		cache: cache,
	}
}

// CacheResponse middleware que cacheia a respostas HTTP
func (cm *CacheMiddleware) CacheResponse(ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Gerar chave de cache baseada na URL e query params
		cacheKey := cm.generateCacheKey(c)

		// Tentar recuperar do cache
		ctx := c.Request.Context()
		cached, err := cm.cache.Get(ctx, cacheKey)
		if err == nil {
			// Cache hit - retornar resposta cached
			c.Header("X-Cache", "HIT")
			c.Data(http.StatusOK, "application/json", []byte(cached))
			c.Abort()
			return
		}

		// Cache miss - continuar processamento
		c.Header("X-Cache", "MISS")

		// Interceptar resposta
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		c.Next()

		// Se status for 200, cachear a resposta
		if writer.statusCode == http.StatusOK && writer.body.Len() > 0 {
			cm.cache.Set(ctx, cacheKey, writer.body.Bytes(), ttl)
		}
	}
}

// generateCacheKey gera uma chave única baseada na requisição
func (cm *CacheMiddleware) generateCacheKey(c *gin.Context) string {
	// Incluir método, path e quey params
	key := fmt.Sprintf("%s:%s?%s", c.Request.Method, c.Request.URL.Path, c.Request.URL.RawQuery)

	// Hash para manter chave pequena
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("cache:http:%s", hex.EncodeToString(hash[:]))
}

// responseWriter intecepta a resposta HTTP
type responseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *responseWriter) Writer(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// InvalidateCache middleware helper para invalidar cache
func (cm *CacheMiddleware) InvalidateCache(pattern string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Invalidar cache após operação de escrita
		ctx := c.Request.Context()
		if err := cm.cache.DeletePattern(ctx, pattern); err != nil {
			// Log erro mas não falha a requisição
			fmt.Printf("Failed to invalidate cache: %v\n", err)
		}
	}
}
