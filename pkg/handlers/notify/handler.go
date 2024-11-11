package notify

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/buglloc/rip/v2/pkg/cfg"
	"github.com/buglloc/rip/v2/pkg/handlers"
	"github.com/buglloc/rip/v2/pkg/hub"
)

const ShortName = "n"
const Name = "notify"

var _ handlers.Handler = (*Handler)(nil)

type Handler struct {
	handlers.BaseHandler
	channel string
	Nested  handlers.Handler
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Name() string {
	return Name
}

func (h *Handler) Init(p handlers.Parser) error {
	init := func() error {
		var err error
		h.channel, err = p.NextRaw()
		if err != nil {
			return err
		}

		h.Nested, err = p.NextHandler()
		if err != nil {
			return err
		}

		return nil
	}

	err := init()
	if err != nil {
		var question = handlers.Question{RemoteAddr: p.RemoteAddr(), Question: dns.Question{Name: p.FQDN()}}
		h.reportErr(question, fmt.Sprintf("can't parse request: %v", err))
	}

	return err
}

func (h *Handler) Handle(question handlers.Question) ([]dns.RR, bool, error) {
	rr, moveOn, err := h.Nested.Handle(question)
	if cfg.HubEnabled {
		if err != nil {
			h.reportErr(question, err.Error())
		} else {
			h.reportRR(question, rr)
		}
	}

	return rr, moveOn, err
}

func (h *Handler) reportRR(question handlers.Question, rr []dns.RR) {
	if h.channel == "" {
		return
	}

	var remoteIP net.IP
	var remotePort int
	var remoteNetwork string

	switch addr := question.RemoteAddr.(type) {
	case *net.UDPAddr:
		remoteIP = addr.IP
		remotePort = addr.Port
		remoteNetwork = addr.Network()
	case *net.TCPAddr:
		remoteIP = addr.IP
		remotePort = addr.Port
		remoteNetwork = addr.Network()
	}

	now := time.Now()
	if len(rr) == 0 {
		hub.Send(h.channel, hub.Message{
			Time:          now,
			RemoteIP:      remoteIP,
			RemotePort:    remotePort,
			RemoteNetwork: remoteNetwork,
			Name:          question.Name,
			QType:         dns.Type(question.Qtype).String(),
			RR:            "<empty>",
			Ok:            true,
		})
		return
	}

	for _, r := range rr {
		hub.Send(h.channel, hub.Message{
			Time:          now,
			RemoteIP:      remoteIP,
			RemotePort:    remotePort,
			RemoteNetwork: remoteNetwork,
			Name:          question.Name,
			QType:         dns.Type(question.Qtype).String(),
			RR:            strings.TrimPrefix(r.String(), r.Header().String()),
			Ok:            true,
		})
	}
}

func (h *Handler) reportErr(question handlers.Question, err string) {
	if h.channel == "" {
		return
	}

	var remoteIP net.IP
	var remotePort int
	var remoteNetwork string

	switch addr := question.RemoteAddr.(type) {
	case *net.UDPAddr:
		remoteIP = addr.IP
		remotePort = addr.Port
		remoteNetwork = addr.Network()
	case *net.TCPAddr:
		remoteIP = addr.IP
		remotePort = addr.Port
		remoteNetwork = addr.Network()
	}

	qType := "n/a"
	if question.Qtype != dns.TypeNone {
		qType = dns.Type(question.Qtype).String()
	}

	hub.Send(h.channel, hub.Message{
		Time:          time.Now(),
		RemoteIP:      remoteIP,
		RemotePort:    remotePort,
		RemoteNetwork: remoteNetwork,
		Name:          question.Name,
		QType:         qType,
		RR:            err,
		Ok:            false,
	})
}
