package ovh

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/libdns/ovh"
)

func parse(t *testing.T, input string) (*Provider, error) {
	t.Helper()
	p := &Provider{new(ovh.Provider)}
	return p, p.UnmarshalCaddyfile(caddyfile.NewTestDispenser(input))
}

func TestUnmarshalCaddyfileOAuth2(t *testing.T) {
	p, err := parse(t, `ovh {
		endpoint ovh-eu
		client_id id-value
		client_secret secret-value
	}`)
	if err != nil {
		t.Fatalf("UnmarshalCaddyfile: %v", err)
	}
	if p.Endpoint != "ovh-eu" || p.ClientID != "id-value" || p.ClientSecret != "secret-value" {
		t.Fatalf("parsed endpoint=%q client_id=%q client_secret=%q", p.Endpoint, p.ClientID, p.ClientSecret)
	}
}

func TestUnmarshalCaddyfileRejectsExtraArgument(t *testing.T) {
	if _, err := parse(t, `ovh {
		client_id a b
	}`); err == nil {
		t.Fatal("client_id with two arguments was accepted")
	}
}

func TestProvisionReplacesFilePlaceholders(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	idFile := write("OVH_CLIENT_ID", "EU.id-from-file\n")
	secretFile := write("OVH_CLIENT_SECRET", "secret-from-file\n")

	p, err := parse(t, `ovh {
		endpoint ovh-eu
		client_id {file.`+idFile+`}
		client_secret {file.`+secretFile+`}
	}`)
	if err != nil {
		t.Fatalf("UnmarshalCaddyfile: %v", err)
	}
	if err := p.Provision(caddy.Context{}); err != nil {
		t.Fatalf("Provision: %v", err)
	}
	if p.ClientID != "EU.id-from-file" || p.ClientSecret != "secret-from-file" {
		t.Fatalf("after Provision: client_id=%q client_secret=%q", p.ClientID, p.ClientSecret)
	}
}
