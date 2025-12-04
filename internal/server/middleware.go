package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware configura CORS (Cross-Origin Resource Sharing) para a API
// Permite que aplicações frontend em diferentes origens façam requisições à API
// Importante para SPAs (Single Page Applications) hospedadas em domínios diferentes
func CORSMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Permitir requisições de qualquer origem (*)
		// Em produção, substituir por origem específica do frontend
		ctx.Header("Access-Control-Allow-Origin", "*")
		// Permitir envio de credenciais (cookies, headers de autenticação)
		ctx.Header("Access-Control-Allow-Credentials", "true")
		// Headers permitidos nas requisições
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		// Métodos HTTP permitidos
		ctx.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// Requisições OPTIONS são preflight requests do CORS
		// Responder com 204 No Content para permitir a requisição real
		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Continuar para o próximo handler
		ctx.Next()
	}
}

// LogginMiddleware cria um middleware de logging customizado
// Formata os logs de requisições HTTP no formato Apache Common Log Format
func LogginMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		// Formato: IP - [timestamp] "METHOD PATH PROTO" STATUS LATENCY "USER_AGENT" ERROR
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			params.ClientIP,                                    // IP do cliente
			params.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"), // Timestamp formatado
			params.Method,                                     // Método HTTP (GET, POST, etc)
			params.Path,                                       // Caminho da requisição
			params.Request.Proto,                              // Protocolo HTTP (HTTP/1.1, etc)
			params.StatusCode,                                 // Código de status HTTP
			params.Latency,                                    // Tempo de processamento
			params.Request.UserAgent(),                         // User-Agent do cliente
			params.ErrorMessage,                               // Mensagem de erro (se houver)
		)
	})
}
