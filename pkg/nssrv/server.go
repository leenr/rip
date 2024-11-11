package nssrv

import (
	"context"
	"net"
	"strings"
	"time"

	log "github.com/buglloc/simplelog"
	"github.com/karlseguin/ccache/v3"
	"github.com/miekg/dns"
	"github.com/pires/go-proxyproto"
	"golang.org/x/sync/errgroup"

	"github.com/buglloc/rip/v2/pkg/cfg"
	"github.com/buglloc/rip/v2/pkg/www"
)

type NSSrv struct {
	tcpServer  *dns.Server
	udpServer  *dns.Server
	unixServer *dns.Server
	wwwServer  *www.HttpSrv
	cache      *ccache.Cache[*cachedHandler]
}

func NewSrv() (*NSSrv, error) {
	tcpDnsListener, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return nil, err
	}

	proxyTcpDnsListener := &proxyproto.Listener{
		Listener:          tcpDnsListener,
		ReadHeaderTimeout: 5 * time.Second,
	}

	srv := &NSSrv{
		tcpServer: &dns.Server{
			Addr:         cfg.Addr,
			Listener:     proxyTcpDnsListener,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
		udpServer: &dns.Server{
			Addr:         cfg.Addr,
			Net:          "udp",
			UDPSize:      65535,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},

		cache: ccache.New(ccache.Configure[*cachedHandler]().MaxSize(cfg.CacheSize)),
	}

	srv.udpServer.Handler = srv.newDNSRouter()
	srv.tcpServer.Handler = srv.newDNSRouter()
	if cfg.HubEnabled {
		srv.wwwServer = www.NewHttpSrv()
	}

	return srv, nil
}

func (s *NSSrv) ListenAndServe() error {
	var g errgroup.Group
	g.Go(func() error {
		log.Info("starting TCP-server", "addr", s.tcpServer.Addr)
		err := s.tcpServer.ActivateAndServe()
		if err != nil {
			log.Error("can't start TCP-server", "err", err)
		}
		return err
	})

	g.Go(func() error {
		log.Info("starting UDP-server", "addr", s.udpServer.Addr)
		err := s.udpServer.ListenAndServe()
		if err != nil {
			log.Error("can't start UDP-server", "err", err)
		}
		return err
	})

	if s.wwwServer != nil {
		g.Go(func() error {
			log.Info("starting HTTP-server", "addr", s.wwwServer.Addr())
			err := s.wwwServer.ListenAndServe()
			if err != nil {
				log.Error("can't start HTTP-server", "err", err)
			}
			return err
		})
	}

	return g.Wait()
}

func (s *NSSrv) Shutdown(ctx context.Context) error {
	var g errgroup.Group
	g.Go(func() error {
		return s.tcpServer.ShutdownContext(ctx)
	})

	g.Go(func() error {
		return s.udpServer.ShutdownContext(ctx)
	})

	if s.wwwServer != nil {
		g.Go(func() error {
			return s.wwwServer.Shutdown(ctx)
		})
	}

	done := make(chan error)
	go func() {
		defer close(done)
		done <- g.Wait()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func (s *NSSrv) newDNSRouter() *dns.ServeMux {
	out := dns.NewServeMux()
	for _, zone := range cfg.Zones {
		if !strings.HasSuffix(zone, ".") {
			zone += "."
		}

		out.HandleFunc(zone, func(zone string) func(w dns.ResponseWriter, req *dns.Msg) {
			return func(w dns.ResponseWriter, req *dns.Msg) {
				defer func() { _ = w.Close() }()

				remoteAddr := w.RemoteAddr()
				l := log.Child("client", remoteAddr.String())
				msg := s.handleRequest(zone, req, remoteAddr, &l)
				_ = w.WriteMsg(msg)
			}
		}(zone))
	}
	return out
}
