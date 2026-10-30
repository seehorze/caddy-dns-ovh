package ovh

import (
	"fmt"

	"github.com/libdns/ovh"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

// Provider wraps the provider implementation as a Caddy module.
type Provider struct{ *ovh.Provider }

func init() {
	caddy.RegisterModule(Provider{})
}

// CaddyModule returns the Caddy module information.
func (Provider) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "dns.providers.ovh",
		New: func() caddy.Module {
			return &Provider{new(ovh.Provider)}
		},
	}
}

// Provision implements the Provisioner interface to initialize the OVH client
func (p *Provider) Provision(ctx caddy.Context) error {
	repl := caddy.NewReplacer()
	p.Provider.Endpoint = repl.ReplaceAll(p.Provider.Endpoint, "")
	p.Provider.ApplicationKey = repl.ReplaceAll(p.Provider.ApplicationKey, "")
	p.Provider.ApplicationSecret = repl.ReplaceAll(p.Provider.ApplicationSecret, "")
	p.Provider.ConsumerKey = repl.ReplaceAll(p.Provider.ConsumerKey, "")
	p.Provider.ClientID = repl.ReplaceAll(p.Provider.ClientID, "")
	p.Provider.ClientSecret = repl.ReplaceAll(p.Provider.ClientSecret, "")

	return p.checkCredentials()
}

// checkCredentials refuses an incomplete or ambiguous configuration, so that
// go-ovh never falls back to OVH_* environment variables or ovh.conf files.
func (p *Provider) checkCredentials() error {
	oauth2 := p.Provider.ClientID != "" || p.Provider.ClientSecret != ""
	legacy := p.Provider.ApplicationKey != "" || p.Provider.ApplicationSecret != "" || p.Provider.ConsumerKey != ""

	switch {
	case p.Provider.Endpoint == "":
		return fmt.Errorf("ovh: endpoint is required")
	case oauth2 && legacy:
		return fmt.Errorf("ovh: set client_id/client_secret or application_key/application_secret/consumer_key, not both")
	case oauth2 && (p.Provider.ClientID == "" || p.Provider.ClientSecret == ""):
		return fmt.Errorf("ovh: client_id and client_secret must both be set")
	case legacy && (p.Provider.ApplicationKey == "" || p.Provider.ApplicationSecret == "" || p.Provider.ConsumerKey == ""):
		return fmt.Errorf("ovh: application_key, application_secret and consumer_key must all be set")
	case !oauth2 && !legacy:
		return fmt.Errorf("ovh: no credentials: set client_id and client_secret, or application_key, application_secret and consumer_key")
	}

	return nil
}

// UnmarshalCaddyfile sets up the DNS provider from Caddyfile tokens. Syntax:
//
// ovh {
//     endpoint <string>
//     application_key <string>
//     application_secret <string>
//     consumer_key <string>
// }
//
// or, with an OAuth2 service account:
//
// ovh {
//     endpoint <string>
//     client_id <string>
//     client_secret <string>
// }
//
func (p *Provider) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if d.NextArg() {
			return d.ArgErr()
		}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "endpoint":
				if d.NextArg() {
					p.Provider.Endpoint = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "application_key":
				if d.NextArg() {
					p.Provider.ApplicationKey = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "application_secret":
				if d.NextArg() {
					p.Provider.ApplicationSecret = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "consumer_key":
				if d.NextArg() {
					p.Provider.ConsumerKey = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "client_id":
				if d.NextArg() {
					p.Provider.ClientID = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			case "client_secret":
				if d.NextArg() {
					p.Provider.ClientSecret = d.Val()
				}
				if d.NextArg() {
					return d.ArgErr()
				}
			default:
				return d.Errf("unrecognized subdirective '%s'", d.Val())
			}
		}
	}

	return nil
}

// Interface guards
var (
	_ caddyfile.Unmarshaler = (*Provider)(nil)
	_ caddy.Provisioner     = (*Provider)(nil)
)