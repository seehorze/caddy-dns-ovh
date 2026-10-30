package ovh

import (
	"os"
	"path/filepath"
	"strings"
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

func TestProvisionCredentialRules(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "absent")
	for _, tc := range []struct {
		name    string
		p       *ovh.Provider
		wantErr string
	}{
		{"oauth2", &ovh.Provider{Endpoint: "ovh-eu", ClientID: "id", ClientSecret: "secret"}, ""},
		{"legacy", &ovh.Provider{Endpoint: "ovh-eu", ApplicationKey: "ak", ApplicationSecret: "as", ConsumerKey: "ck"}, ""},
		{"no credential", &ovh.Provider{Endpoint: "ovh-eu"}, "no credentials"},
		{"missing secret file", &ovh.Provider{Endpoint: "ovh-eu", ClientID: "{file." + missing + "}", ClientSecret: "{file." + missing + "}"}, "no credentials"},
		{"client id only", &ovh.Provider{Endpoint: "ovh-eu", ClientID: "id"}, "client_id and client_secret"},
		{"legacy without consumer key", &ovh.Provider{Endpoint: "ovh-eu", ApplicationKey: "ak", ApplicationSecret: "as"}, "application_key, application_secret and consumer_key"},
		{"both kinds", &ovh.Provider{Endpoint: "ovh-eu", ClientID: "id", ClientSecret: "secret", ConsumerKey: "ck"}, "not both"},
		{"no endpoint", &ovh.Provider{ClientID: "id", ClientSecret: "secret"}, "endpoint"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := (&Provider{tc.p}).Provision(caddy.Context{})
			switch {
			case tc.wantErr == "" && err != nil:
				t.Fatalf("Provision: %v", err)
			case tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)):
				t.Fatalf("Provision error = %v, want it to contain %q", err, tc.wantErr)
			}
		})
	}
}
