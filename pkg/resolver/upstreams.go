package resolver

import (
	"net"
	"time"

	"github.com/miekg/dns"

	"github.com/buglloc/rip/v2/pkg/cfg"
)

var (
	dnsClient = &dns.Client{
		Net:          "tcp",
		ReadTimeout:  time.Second * 1,
		WriteTimeout: time.Second * 1,
	}
	dnsCache = NewCache()
)

func ResolveIp(reqType uint16, name string) ([]net.IP, error) {
	if ips := dnsCache.Get(reqType, name); ips != nil {
		return ips, nil
	}

	msg := &dns.Msg{}
	msg.SetQuestion(dns.Fqdn(name), reqType)
	res, _, err := dnsClient.Exchange(msg, cfg.Upstream)
	if err != nil || len(res.Answer) == 0 {
		return nil, err
	}

	var addresses []net.IP
	for _, rr := range res.Answer {
		switch v := rr.(type) {
		case *dns.A:
			addresses = append(addresses, v.A)
		case *dns.AAAA:
			addresses = append(addresses, v.AAAA)
		}
	}

	ttl := time.Duration(res.Answer[0].Header().Ttl) * time.Second
	dnsCache.Set(reqType, name, ttl, addresses)

	return addresses, nil
}
