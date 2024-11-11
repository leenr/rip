package handlers

import (
	"github.com/miekg/dns"
	"net"

	"github.com/buglloc/rip/v2/pkg/handlers/limiter"
)

type Question struct {
	dns.Question
	RemoteAddr net.Addr
}

type Handler interface {
	Name() string
	Init(p Parser) error
	SetDefaultLimiters(modifiers ...limiter.Limiter)
	Handle(question Question) (rrs []dns.RR, moveOn bool, err error)
}

type BaseHandler struct {
	Limiters limiter.Limiters
}

func (h *BaseHandler) SetDefaultLimiters(modifiers ...limiter.Limiter) {
	if len(h.Limiters) == 0 {
		h.Limiters = modifiers
	}
}
