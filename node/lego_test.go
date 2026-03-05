package node

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MoeclubM/V2bX/conf"
)

func newIntegrationLego(t *testing.T) *Lego {
	if os.Getenv("V2BX_RUN_ACME_DNS_TESTS") != "1" {
		t.Skip("skip ACME DNS integration test, set V2BX_RUN_ACME_DNS_TESTS=1 to enable")
	}
	apiToken := os.Getenv("CF_DNS_API_TOKEN")
	if apiToken == "" {
		t.Skip("skip ACME DNS integration test, missing CF_DNS_API_TOKEN")
	}
	certDomain := os.Getenv("V2BX_TEST_CERT_DOMAIN")
	if certDomain == "" {
		t.Skip("skip ACME DNS integration test, missing V2BX_TEST_CERT_DOMAIN")
	}
	email := os.Getenv("V2BX_TEST_CERT_EMAIL")
	if email == "" {
		email = "test@test.com"
	}
	workDir := t.TempDir()
	l, err := NewLego(&conf.CertConfig{
		CertMode:   "dns",
		Email:      email,
		CertDomain: certDomain,
		Provider:   "cloudflare",
		DNSEnv: map[string]string{
			"CF_DNS_API_TOKEN": apiToken,
		},
		CertFile: filepath.Join(workDir, "cert", "1.pem"),
		KeyFile:  filepath.Join(workDir, "cert", "1.key"),
	})
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestLego_CreateCertByDns(t *testing.T) {
	l := newIntegrationLego(t)
	err := l.CreateCert()
	if err != nil {
		t.Error(err)
	}
}

func TestLego_RenewCert(t *testing.T) {
	l := newIntegrationLego(t)
	if err := l.CreateCert(); err != nil {
		t.Fatal(err)
	}
	if err := l.RenewCert(); err != nil {
		t.Fatal(err)
	}
}
