package loopback

import (
	"fmt"
	"net"

	"github.com/miekg/dns"

	"github.com/buglloc/rip/v2/pkg/handlers"
	"github.com/buglloc/rip/v2/pkg/handlers/limiter"
)

const ShortName = "lo"
const Name = "loopback"

var _ handlers.Handler = (*Handler)(nil)

type Handler struct {
	handlers.BaseHandler
}

func NewHandler(modifiers ...limiter.Limiter) *Handler {
	return &Handler{
		BaseHandler: handlers.BaseHandler{
			Limiters: modifiers,
		},
	}
}

func (h *Handler) Name() string {
	return Name
}

func (h *Handler) Init(p handlers.Parser) error {
	return nil
}

func (h *Handler) Handle(question handlers.Question) ([]dns.RR, bool, error) {
	var remoteIP net.IP

	switch addr := question.RemoteAddr.(type) {
	case *net.UDPAddr:
		remoteIP = addr.IP
	case *net.TCPAddr:
		remoteIP = addr.IP
	default:
		return nil, false, fmt.Errorf("cannot determine remote address")
	}

	// if len(h.IP.To4()) == net.IPv4len {
	//	 if question.Qtype != dns.TypeA && question.Qtype != dns.TypeANY {
	//		 return nil, false, nil
	//	 }
	// } else if len(h.IP) == net.IPv6len {
	//	 if question.Qtype != dns.TypeAAAA && question.Qtype != dns.TypeANY {
	//		 return nil, false, nil
	//	 }
	// }

	h.Limiters.Use()
	return handlers.IPsToRR(question.Question, remoteIP), h.Limiters.MoveOn(), nil
}
