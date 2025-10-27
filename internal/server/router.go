package server

import "github.com/GabrielFerrarez19/gofinance-api/internal/user"

type Router struct {
	userHandler *user.Handler
}

func NewRoute(userHandler *user.Handler) *Router {
	return &Router{
		userHandler: userHandler,
	}
}
