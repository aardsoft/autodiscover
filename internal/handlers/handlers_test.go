package handlers

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aardsoft/autodiscover/internal/config"
)

func testSettings() *config.Settings {
	s := &config.Settings{}
	s.Info.Name = "Test Corp"
	s.Info.URL = "https://test.example.com"
	s.Info.Domain = "test.example.com"
	s.Server.IMAP.Host = "imap.test.example.com"
	s.Server.IMAP.Port = 993
	s.Server.IMAP.Socket = "SSL"
	s.Server.SMTP.Host = "smtp.test.example.com"
	s.Server.SMTP.Port = 465
	s.Server.SMTP.Socket = "SSL"
	s.Server.DomainRequired = true
	s.TTL = 168

	s.CalDAV = &config.CalDAVSettings{
		Host:         "dav.test.example.com",
		Port:         443,
		SSL:          true,
		PrincipalURL: "/dav/calendars/user/%EMAILADDRESS%",
	}
	s.CardDAV = &config.CardDAVSettings{
		Host:         "dav.test.example.com",
		Port:         443,
		SSL:          true,
		PrincipalURL: "/dav/addressbooks/user/%EMAILADDRESS%",
	}
	s.Exchange = &config.ExchangeSettings{
		EWSURL: "https://mail.test.example.com/ews/exchange.asmx",
	}
	s.ActiveSync = &config.ActiveSyncSettings{
		URL: "https://mail.test.example.com/Microsoft-Server-ActiveSync",
	}
	s.VPN = &config.VPNSettings{
		Type:     "IKEv2",
		Host:     "vpn.test.example.com",
		RemoteID: "vpn.test.example.com",
	}
	s.MDM = &config.MDMSettings{
		ServerURL: "https://mdm.test.example.com",
	}
	return s
}

func TestMSAutodiscoverOutlook(t *testing.T) {
	s := testSettings()
	body := `<?xml version="1.0" encoding="utf-8"?>
<Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/outlook/requestschema/2006">
  <Request>
    <EMailAddress>user@test.example.com</EMailAddress>
    <AcceptableResponseSchema>http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a</AcceptableResponseSchema>
  </Request>
</Autodiscover>`

	req := httptest.NewRequest(http.MethodPost, "/autodiscover/autodiscover.xml", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "text/xml")
	w := httptest.NewRecorder()

	HandleMSAutodiscover(s)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := w.Body.String()
	if !strings.Contains(resp, `<Type>IMAP</Type>`) {
		t.Error("response missing IMAP protocol")
	}
	if !strings.Contains(resp, `<Type>SMTP</Type>`) {
		t.Error("response missing SMTP protocol")
	}
	if !strings.Contains(resp, `<Type>EXPR</Type>`) {
		t.Error("response missing EXPR protocol")
	}
	if !strings.Contains(resp, `<Server>imap.test.example.com</Server>`) {
		t.Error("response missing IMAP server")
	}
	if strings.Contains(resp, `<Server></Server>`) {
		t.Error("response contains empty Server tag")
	}
}

func TestMSAutodiscoverActiveSync(t *testing.T) {
	s := testSettings()
	body := `<?xml version="1.0" encoding="utf-8"?>
<Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/mobilesync/requestschema/2006">
  <Request>
    <EMailAddress>user@test.example.com</EMailAddress>
    <AcceptableResponseSchema>http://schemas.microsoft.com/exchange/autodiscover/mobilesync/responseschema/2006</AcceptableResponseSchema>
  </Request>
</Autodiscover>`

	req := httptest.NewRequest(http.MethodPost, "/autodiscover/autodiscover.xml", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "text/xml")
	w := httptest.NewRecorder()

	HandleMSAutodiscover(s)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := w.Body.String()
	if !strings.Contains(resp, `<Type>MobileSync</Type>`) {
		t.Error("response missing MobileSync type")
	}
	if !strings.Contains(resp, `https://mail.test.example.com/Microsoft-Server-ActiveSync`) {
		t.Error("response missing ActiveSync URL")
	}
}

func TestThunderbirdAutoconfig(t *testing.T) {
	s := testSettings()
	req := httptest.NewRequest(http.MethodGet, "/mail/config-v1.1.xml?emailaddress=user@test.example.com", nil)
	w := httptest.NewRecorder()

	HandleThunderbird(s)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := w.Body.String()
	if !strings.Contains(resp, `<incomingServer type="imap">`) {
		t.Error("response missing IMAP incoming server")
	}
	if !strings.Contains(resp, `<outgoingServer type="smtp">`) {
		t.Error("response missing SMTP outgoing server")
	}
	if !strings.Contains(resp, `<addressBook type="carddav">`) {
		t.Error("response missing CardDAV address book")
	}
	if !strings.Contains(resp, `<calendar type="caldav">`) {
		t.Error("response missing CalDAV calendar")
	}
	if !strings.Contains(resp, `%EMAILADDRESS%`) {
		t.Error("response missing email placeholders")
	}

	// Validate XML structure
	var cfg thunderbirdConfig
	if err := xml.Unmarshal(w.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("failed to unmarshal thunderbird config: %v", err)
	}
	if cfg.Version != "1.1" {
		t.Errorf("expected version 1.1, got %s", cfg.Version)
	}
	if cfg.EmailProvider.ID != "test.example.com" {
		t.Errorf("expected provider id test.example.com, got %s", cfg.EmailProvider.ID)
	}
}

func TestAppleProfile(t *testing.T) {
	s := testSettings()
	req := httptest.NewRequest(http.MethodGet, "/profile.mobileconfig?email=user@test.example.com", nil)
	w := httptest.NewRecorder()

	HandleAppleProfile(s)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/x-apple-aspen-config" {
		t.Errorf("expected Content-Type application/x-apple-aspen-config, got %s", ct)
	}

	resp := w.Body.String()
	if !strings.Contains(resp, `<key>PayloadType</key>`) {
		t.Error("response missing PayloadType keys")
	}
	if !strings.Contains(resp, `com.apple.mail.managed`) {
		t.Error("response missing email payload")
	}
	if !strings.Contains(resp, `com.apple.caldav.account`) {
		t.Error("response missing caldav payload")
	}
	if !strings.Contains(resp, `com.apple.carddav.account`) {
		t.Error("response missing carddav payload")
	}
	if !strings.Contains(resp, `com.apple.vpn.managed`) {
		t.Error("response missing vpn payload")
	}
	if !strings.Contains(resp, `com.apple.mdm`) {
		t.Error("response missing mdm payload")
	}
	if !strings.Contains(resp, `user@test.example.com`) {
		t.Error("response missing expanded email address")
	}
}

func TestWellKnownCalDAV(t *testing.T) {
	s := testSettings()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/caldav?email=user@test.example.com", nil)
	w := httptest.NewRecorder()

	HandleWellKnownCalDAV(s)(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", w.Code)
	}

	loc := w.Header().Get("Location")
	expected := "https://dav.test.example.com/dav/calendars/user/user@test.example.com"
	if loc != expected {
		t.Errorf("expected redirect to %s, got %s", expected, loc)
	}
}

func TestWellKnownCardDAV(t *testing.T) {
	s := testSettings()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/carddav?email=user@test.example.com", nil)
	w := httptest.NewRecorder()

	HandleWellKnownCardDAV(s)(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", w.Code)
	}

	loc := w.Header().Get("Location")
	expected := "https://dav.test.example.com/dav/addressbooks/user/user@test.example.com"
	if loc != expected {
		t.Errorf("expected redirect to %s, got %s", expected, loc)
	}
}

func TestEmailLocalPart(t *testing.T) {
	cases := []struct{ in, want string }{
		{"user@example.com", "user"},
		{"a@b.c", "a"},
		{"nope", "nope"},
		{"@example.com", "@example.com"},
	}
	for _, c := range cases {
		got := emailLocalPart(c.in)
		if got != c.want {
			t.Errorf("emailLocalPart(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestEmailDomain(t *testing.T) {
	cases := []struct{ in, want string }{
		{"user@example.com", "example.com"},
		{"a@b.c", "b.c"},
		{"nope", ""},
		{"@example.com", "example.com"},
	}
	for _, c := range cases {
		got := emailDomain(c.in)
		if got != c.want {
			t.Errorf("emailDomain(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSocketTypeToSSL(t *testing.T) {
	if !socketTypeToSSL("SSL") {
		t.Error("expected SSL socket to return true")
	}
	if socketTypeToSSL("STARTTLS") {
		t.Error("expected STARTTLS socket to return false")
	}
}
