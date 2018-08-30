package commands

import (
	"errors"
	"net"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/buglloc/rip/pkg/cfg"
	"github.com/buglloc/rip/pkg/cli"
	"github.com/buglloc/rip/pkg/ns_server"
)

var nsServerCmd = &cobra.Command{
	Use:     "ns --zone=example.com --zone=example1.com",
	Short:   "Start RIP NS server",
	PreRunE: parseServerConfig,
	RunE:    runServerCmd,
}

func init() {
	flags := nsServerCmd.PersistentFlags()
	flags.String("listen", ":53",
		"address to listen on")
	flags.StringSlice("zone", []string{"."},
		"your zone name (e.g. 'buglloc.com')")
	flags.String("ipv4", "127.0.0.1",
		"default ipv4 address")
	flags.Uint32("ttl", cfg.TTL,
		"DNS records TTL")
	flags.Uint32("sticky-ttl", 30,
		"sticky record TTL in seconds")
	flags.String("ipv6", "::1",
		"default ipv6 address")
	flags.String("upstream", "77.88.8.8:53",
		"upstream DNS server")
	flags.Bool("use-default", false,
		"return default IPs for not supported requests")
	flags.Bool("no-proxy", false,
		"disable proxy mode")

	cli.BindPFlags(flags)
	RootCmd.AddCommand(nsServerCmd)
}

func runServerCmd(cmd *cobra.Command, args []string) error {
	ns_server.RunBackground()
	cli.ListenInterrupt()
	return nil
}

func parseServerConfig(cmd *cobra.Command, args []string) error {
	cfg.Zones = viper.GetStringSlice("Zone")
	if len(cfg.Zones) == 0 {
		return errors.New("empty zone list, please provide at leas one")
	}

	cfg.Addr = viper.GetString("Listen")
	cfg.IPv4 = net.ParseIP(viper.GetString("Ipv4"))
	cfg.IPv6 = net.ParseIP(viper.GetString("Ipv6"))
	cfg.AllowProxy = !viper.GetBool("NoProxy")
	cfg.UseDefault = viper.GetBool("UseDefault")
	cfg.Upstream = viper.GetString("Upstream")
	cfg.TTL = uint32(viper.GetInt("Ttl"))
	cfg.StickyTTL = time.Duration(viper.GetInt("StickyTtl")) * time.Second
	return nil
}
