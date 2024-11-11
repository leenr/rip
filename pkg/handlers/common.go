package handlers

import (
	"net"
	"strings"

	"github.com/miekg/dns"

	"github.com/buglloc/rip/v2/pkg/cfg"
	"github.com/buglloc/rip/v2/pkg/iputil"
)

func PartToIP(part string) net.IP {
	dotCounts := strings.Count(part, "-")

	switch dotCounts {
	case 0:
		ip, _ := iputil.DecodeIp(part)
		return ip
	case 3:
		return net.ParseIP(strings.ReplaceAll(part, "-", ".")).To4()
	default:
		return net.ParseIP(strings.ReplaceAll(part, "-", ":")).To16()
	}
}

func DefaultIp(reqType uint16) net.IP {
	if reqType == dns.TypeA {
		return cfg.IPv4
	}
	return cfg.IPv6
}

func IPsToRR(question dns.Question, ips ...net.IP) (result []dns.RR) {
	result = make([]dns.RR, len(ips))
	for i, ip := range ips {
		result[i] = createIpRR(question, ip)
	}

	return
}

func createIpRR(question dns.Question, ip net.IP) (rr dns.RR) {
	if ip.To4() != nil {
		rr = &dns.A{
			Hdr: dns.RR_Header{
				Name:   question.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    cfg.TTL,
			},
			A: ip.To4(),
		}
	} else if len(ip) == net.IPv6len {
		rr = &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   question.Name,
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    cfg.TTL,
			},
			AAAA: ip.To16(),
		}
	}

	return
}
