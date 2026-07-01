package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/aardsoft/autodiscover/internal/config"
)

// HandleAppleProfile serves a signed Apple configuration profile (.mobileconfig).
func HandleAppleProfile(s *config.Settings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := strings.TrimSpace(r.URL.Query().Get("email"))
		if email == "" {
			email = "user@" + s.Info.Domain
		}

		data := buildMobileConfig(s, email)
		writeMobileConfig(w, http.StatusOK, []byte(data))
	}
}

// HandleWellKnownCalDAV redirects /.well-known/caldav to the configured CalDAV principal URL.
func HandleWellKnownCalDAV(s *config.Settings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.CalDAV == nil || s.CalDAV.Host == "" {
			handleNotConfigured(w, "CalDAV")
			return
		}
		scheme := "http"
		if s.CalDAV.SSL {
			scheme = "https"
		}
		email := strings.TrimSpace(r.URL.Query().Get("email"))
		if email == "" {
			email = "user@" + s.Info.Domain
		}
		path := expandPlaceholders(s.CalDAV.PrincipalURL, email)
		url := fullURL(scheme, s.CalDAV.Host, s.CalDAV.Port, path)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// HandleWellKnownCardDAV redirects /.well-known/carddav to the configured CardDAV principal URL.
func HandleWellKnownCardDAV(s *config.Settings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.CardDAV == nil || s.CardDAV.Host == "" {
			handleNotConfigured(w, "CardDAV")
			return
		}
		scheme := "http"
		if s.CardDAV.SSL {
			scheme = "https"
		}
		email := strings.TrimSpace(r.URL.Query().Get("email"))
		if email == "" {
			email = "user@" + s.Info.Domain
		}
		path := expandPlaceholders(s.CardDAV.PrincipalURL, email)
		url := fullURL(scheme, s.CardDAV.Host, s.CardDAV.Port, path)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func buildMobileConfig(s *config.Settings, email string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">` + "\n")
	b.WriteString(`<plist version="1.0">` + "\n")
	b.WriteString("<dict>\n")
	b.WriteString("\t<key>PayloadContent</key>\n")
	b.WriteString("\t<array>\n")

	// Email payload
	if s.Server.IMAP.Host != "" && s.Server.SMTP.Host != "" {
		b.WriteString(buildEmailPayload(s, email))
	}

	// CalDAV payload
	if s.CalDAV != nil && s.CalDAV.Host != "" {
		b.WriteString(buildCalDAVPayload(s, email))
	}

	// CardDAV payload
	if s.CardDAV != nil && s.CardDAV.Host != "" {
		b.WriteString(buildCardDAVPayload(s, email))
	}

	// VPN payload
	if s.VPN != nil && s.VPN.Host != "" {
		b.WriteString(buildVPNPayload(s, email))
	}

	// MDM payload
	if s.MDM != nil && s.MDM.ServerURL != "" {
		b.WriteString(buildMDMPayload(s, email))
	}

	b.WriteString("\t</array>\n")

	// Root profile keys
	b.WriteString("\t<key>PayloadDisplayName</key>\n")
	b.WriteString(fmt.Sprintf("\t<string>%s</string>\n", xmlEsc(s.Info.Name+" Configuration")))
	b.WriteString("\t<key>PayloadIdentifier</key>\n")
	b.WriteString(fmt.Sprintf("\t<string>%s</string>\n", xmlEsc("com."+s.Info.Domain+".autodiscover")))
	b.WriteString("\t<key>PayloadType</key>\n")
	b.WriteString("\t<string>Configuration</string>\n")
	b.WriteString("\t<key>PayloadUUID</key>\n")
	b.WriteString(fmt.Sprintf("\t<string>%s</string>\n", payloadUUID(s.Info.Domain+"-root")))
	b.WriteString("\t<key>PayloadVersion</key>\n")
	b.WriteString("\t<integer>1</integer>\n")
	b.WriteString("</dict>\n")
	b.WriteString("</plist>\n")

	return b.String()
}

func buildEmailPayload(s *config.Settings, email string) string {
	var b strings.Builder
	b.WriteString("\t\t<dict>\n")
	addKeyString(&b, "PayloadType", "com.apple.mail.managed")
	addKeyString(&b, "PayloadIdentifier", "com."+s.Info.Domain+".email")
	addKeyString(&b, "PayloadUUID", payloadUUID(s.Info.Domain+"-email"))
	addKeyInt(&b, "PayloadVersion", 1)
	addKeyString(&b, "EmailAccountDescription", s.Info.Name+" Mail")
	addKeyString(&b, "EmailAccountName", emailLocalPart(email))
	addKeyString(&b, "EmailAccountType", "EmailTypeIMAP")
	addKeyString(&b, "EmailAddress", email)
	addKeyString(&b, "IncomingMailServerAuthentication", "EmailAuthPassword")
	addKeyString(&b, "IncomingMailServerHostName", s.Server.IMAP.Host)
	addKeyInt(&b, "IncomingMailServerPortNumber", s.Server.IMAP.Port)
	addKeyBool(&b, "IncomingMailServerUseSSL", socketTypeToSSL(s.Server.IMAP.Socket))
	addKeyString(&b, "IncomingMailServerUsername", email)
	addKeyString(&b, "OutgoingMailServerAuthentication", "EmailAuthPassword")
	addKeyString(&b, "OutgoingMailServerHostName", s.Server.SMTP.Host)
	addKeyInt(&b, "OutgoingMailServerPortNumber", s.Server.SMTP.Port)
	addKeyBool(&b, "OutgoingMailServerUseSSL", socketTypeToSSL(s.Server.SMTP.Socket))
	addKeyString(&b, "OutgoingMailServerUsername", email)
	b.WriteString("\t\t</dict>\n")
	return b.String()
}

func buildCalDAVPayload(s *config.Settings, email string) string {
	var b strings.Builder
	b.WriteString("\t\t<dict>\n")
	addKeyString(&b, "PayloadType", "com.apple.caldav.account")
	addKeyString(&b, "PayloadIdentifier", "com."+s.Info.Domain+".caldav")
	addKeyString(&b, "PayloadUUID", payloadUUID(s.Info.Domain+"-caldav"))
	addKeyInt(&b, "PayloadVersion", 1)
	addKeyString(&b, "CalDAVAccountDescription", s.Info.Name+" Calendar")
	addKeyString(&b, "CalDAVHostName", s.CalDAV.Host)
	addKeyInt(&b, "CalDAVPort", s.CalDAV.Port)
	addKeyString(&b, "CalDAVPrincipalURL", expandPlaceholders(s.CalDAV.PrincipalURL, email))
	addKeyBool(&b, "CalDAVUseSSL", s.CalDAV.SSL)
	addKeyString(&b, "CalDAVUsername", email)
	b.WriteString("\t\t</dict>\n")
	return b.String()
}

func buildCardDAVPayload(s *config.Settings, email string) string {
	var b strings.Builder
	b.WriteString("\t\t<dict>\n")
	addKeyString(&b, "PayloadType", "com.apple.carddav.account")
	addKeyString(&b, "PayloadIdentifier", "com."+s.Info.Domain+".carddav")
	addKeyString(&b, "PayloadUUID", payloadUUID(s.Info.Domain+"-carddav"))
	addKeyInt(&b, "PayloadVersion", 1)
	addKeyString(&b, "CardDAVAccountDescription", s.Info.Name+" Contacts")
	addKeyString(&b, "CardDAVHostName", s.CardDAV.Host)
	addKeyInt(&b, "CardDAVPort", s.CardDAV.Port)
	addKeyString(&b, "CardDAVPrincipalURL", expandPlaceholders(s.CardDAV.PrincipalURL, email))
	addKeyBool(&b, "CardDAVUseSSL", s.CardDAV.SSL)
	addKeyString(&b, "CardDAVUsername", email)
	b.WriteString("\t\t</dict>\n")
	return b.String()
}

func buildVPNPayload(s *config.Settings, email string) string {
	var b strings.Builder
	b.WriteString("\t\t<dict>\n")
	addKeyString(&b, "PayloadType", "com.apple.vpn.managed")
	addKeyString(&b, "PayloadIdentifier", "com."+s.Info.Domain+".vpn")
	addKeyString(&b, "PayloadUUID", payloadUUID(s.Info.Domain+"-vpn"))
	addKeyInt(&b, "PayloadVersion", 1)
	addKeyString(&b, "UserDefinedName", s.Info.Name+" VPN")
	addKeyString(&b, "VPNType", s.VPN.Type)

	vpnTypeUpper := strings.ToUpper(s.VPN.Type)
	if vpnTypeUpper == "IKEV2" {
		b.WriteString("\t\t\t<key>IKEv2</key>\n")
		b.WriteString("\t\t\t<dict>\n")
		b.WriteString(fmt.Sprintf("\t\t\t\t<key>RemoteAddress</key>\n"))
		b.WriteString(fmt.Sprintf("\t\t\t\t<string>%s</string>\n", xmlEsc(s.VPN.Host)))
		if s.VPN.RemoteID != "" {
			b.WriteString(fmt.Sprintf("\t\t\t\t<key>RemoteIdentifier</key>\n"))
			b.WriteString(fmt.Sprintf("\t\t\t\t<string>%s</string>\n", xmlEsc(s.VPN.RemoteID)))
		}
		if s.VPN.LocalID != "" {
			b.WriteString(fmt.Sprintf("\t\t\t\t<key>LocalIdentifier</key>\n"))
			b.WriteString(fmt.Sprintf("\t\t\t\t<string>%s</string>\n", xmlEsc(s.VPN.LocalID)))
		}
		if s.VPN.AuthMethod != "" {
			b.WriteString(fmt.Sprintf("\t\t\t\t<key>AuthenticationMethod</key>\n"))
			b.WriteString(fmt.Sprintf("\t\t\t\t<string>%s</string>\n", xmlEsc(s.VPN.AuthMethod)))
		} else {
			b.WriteString(fmt.Sprintf("\t\t\t\t<key>AuthenticationMethod</key>\n"))
			b.WriteString(fmt.Sprintf("\t\t\t\t<string>None</string>\n"))
		}
		b.WriteString("\t\t\t</dict>\n")
	} else {
		// Generic VPN config - just host
		b.WriteString(fmt.Sprintf("\t\t\t<key>%s</key>\n", vpnTypeUpper))
		b.WriteString("\t\t\t<dict>\n")
		b.WriteString(fmt.Sprintf("\t\t\t\t<key>RemoteAddress</key>\n"))
		b.WriteString(fmt.Sprintf("\t\t\t\t<string>%s</string>\n", xmlEsc(s.VPN.Host)))
		b.WriteString("\t\t\t</dict>\n")
	}
	b.WriteString("\t\t</dict>\n")
	return b.String()
}

func buildMDMPayload(s *config.Settings, email string) string {
	var b strings.Builder
	b.WriteString("\t\t<dict>\n")
	addKeyString(&b, "PayloadType", "com.apple.mdm")
	addKeyString(&b, "PayloadIdentifier", "com."+s.Info.Domain+".mdm")
	addKeyString(&b, "PayloadUUID", payloadUUID(s.Info.Domain+"-mdm"))
	addKeyInt(&b, "PayloadVersion", 1)
	addKeyString(&b, "PayloadOrganization", s.Info.Name)
	addKeyString(&b, "ServerURL", s.MDM.ServerURL)
	addKeyBool(&b, "CheckInURL", false) // placeholder
	addKeyBool(&b, "CheckOutWhenRemoved", true)
	addKeyBool(&b, "SignMessage", true)
	addKeyString(&b, "Topic", "com.apple.mgmt."+s.Info.Domain)
	b.WriteString("\t\t</dict>\n")
	return b.String()
}

func addKeyString(b *strings.Builder, key, value string) {
	b.WriteString(fmt.Sprintf("\t\t\t<key>%s</key>\n", xmlEsc(key)))
	b.WriteString(fmt.Sprintf("\t\t\t<string>%s</string>\n", xmlEsc(value)))
}

func addKeyInt(b *strings.Builder, key string, value int) {
	b.WriteString(fmt.Sprintf("\t\t\t<key>%s</key>\n", xmlEsc(key)))
	b.WriteString(fmt.Sprintf("\t\t\t<integer>%d</integer>\n", value))
}

func addKeyBool(b *strings.Builder, key string, value bool) {
	b.WriteString(fmt.Sprintf("\t\t\t<key>%s</key>\n", xmlEsc(key)))
	if value {
		b.WriteString("\t\t\t<true/>\n")
	} else {
		b.WriteString("\t\t\t<false/>\n")
	}
}

func xmlEsc(s string) string {
	return xmlEscapeString(s)
}

func xmlEscapeString(s string) string {
	var b strings.Builder
	if err := xml.EscapeText(&b, []byte(s)); err != nil {
		return s
	}
	return b.String()
}

func payloadUUID(seed string) string {
	h := sha256.New()
	h.Write([]byte(seed))
	return hex.EncodeToString(h.Sum(nil))[:32]
}
