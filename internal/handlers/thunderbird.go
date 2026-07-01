package handlers

import (
	"encoding/xml"
	"net/http"
	"strings"

	"github.com/aardsoft/autodiscover/internal/config"
)

// Thunderbird autoconfig structures
type thunderbirdConfig struct {
	XMLName       xml.Name                 `xml:"clientConfig"`
	Version       string                   `xml:"version,attr"`
	EmailProvider thunderbirdEmailProvider `xml:"emailProvider"`
}

type thunderbirdEmailProvider struct {
	ID               string                      `xml:"id,attr"`
	Domain           []string                    `xml:"domain"`
	DisplayName      string                      `xml:"displayName"`
	DisplayShortName string                      `xml:"displayShortName"`
	IncomingServers  []thunderbirdIncomingServer `xml:"incomingServer"`
	OutgoingServers  []thunderbirdOutgoingServer `xml:"outgoingServer"`
	AddressBooks     []thunderbirdAddressBook    `xml:"addressBook,omitempty"`
	Calendars        []thunderbirdCalendar       `xml:"calendar,omitempty"`
	Documentation    *thunderbirdDocumentation   `xml:"documentation,omitempty"`
}

type thunderbirdIncomingServer struct {
	Type           string `xml:"type,attr"`
	Hostname       string `xml:"hostname"`
	Port           int    `xml:"port"`
	SocketType     string `xml:"socketType"`
	Authentication string `xml:"authentication"`
	Username       string `xml:"username"`
}

type thunderbirdOutgoingServer struct {
	Type           string `xml:"type,attr"`
	Hostname       string `xml:"hostname"`
	Port           int    `xml:"port"`
	SocketType     string `xml:"socketType"`
	Authentication string `xml:"authentication"`
	Username       string `xml:"username"`
}

type thunderbirdAddressBook struct {
	Type     string `xml:"type,attr"`
	URI      string `xml:"uri"`
	Username string `xml:"username"`
}

type thunderbirdCalendar struct {
	Type     string `xml:"type,attr"`
	URI      string `xml:"uri"`
	Username string `xml:"username"`
}

type thunderbirdDocumentation struct {
	URL   string             `xml:"url,attr"`
	Descr []thunderbirdDescr `xml:"descr"`
}

type thunderbirdDescr struct {
	Lang string `xml:"lang,attr"`
	Text string `xml:",chardata"`
}

// HandleThunderbird serves Thunderbird/Mozilla autoconfig XML.
func HandleThunderbird(s *config.Settings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := strings.TrimSpace(r.URL.Query().Get("emailaddress"))
		if email == "" {
			email = strings.TrimSpace(r.URL.Query().Get("email"))
		}
		if email == "" {
			// Use a placeholder if no email provided
			email = "user@" + s.Info.Domain
		}

		cfg := buildThunderbirdConfig(s, email)
		writeXML(w, http.StatusOK, cfg)
	}
}

func buildThunderbirdConfig(s *config.Settings, email string) thunderbirdConfig {
	domain := s.Info.Domain
	if domain == "" {
		domain = emailDomain(email)
	}

	provider := thunderbirdEmailProvider{
		ID:               domain,
		Domain:           []string{domain},
		DisplayName:      s.Info.Name,
		DisplayShortName: s.Info.Name,
	}

	if s.Server.IMAP.Host != "" {
		provider.IncomingServers = append(provider.IncomingServers, thunderbirdIncomingServer{
			Type:           "imap",
			Hostname:       s.Server.IMAP.Host,
			Port:           s.Server.IMAP.Port,
			SocketType:     socketTypeToThunderbirdSocket(s.Server.IMAP.Socket),
			Authentication: "password-cleartext",
			Username:       "%EMAILADDRESS%",
		})
	}

	if s.Server.SMTP.Host != "" {
		provider.OutgoingServers = append(provider.OutgoingServers, thunderbirdOutgoingServer{
			Type:           "smtp",
			Hostname:       s.Server.SMTP.Host,
			Port:           s.Server.SMTP.Port,
			SocketType:     socketTypeToThunderbirdSocket(s.Server.SMTP.Socket),
			Authentication: "password-cleartext",
			Username:       "%EMAILADDRESS%",
		})
	}

	if s.CardDAV != nil && s.CardDAV.Host != "" {
		scheme := "http"
		if s.CardDAV.SSL {
			scheme = "https"
		}
		uri := fullURL(scheme, s.CardDAV.Host, s.CardDAV.Port, s.CardDAV.PrincipalURL)
		provider.AddressBooks = append(provider.AddressBooks, thunderbirdAddressBook{
			Type:     "carddav",
			URI:      uri,
			Username: "%EMAILADDRESS%",
		})
	}

	if s.CalDAV != nil && s.CalDAV.Host != "" {
		scheme := "http"
		if s.CalDAV.SSL {
			scheme = "https"
		}
		uri := fullURL(scheme, s.CalDAV.Host, s.CalDAV.Port, s.CalDAV.PrincipalURL)
		provider.Calendars = append(provider.Calendars, thunderbirdCalendar{
			Type:     "caldav",
			URI:      uri,
			Username: "%EMAILADDRESS%",
		})
	}

	if s.Info.URL != "" {
		provider.Documentation = &thunderbirdDocumentation{
			URL: s.Info.URL,
			Descr: []thunderbirdDescr{
				{Lang: "de", Text: "Allgemeine Beschreibung der Einstellungen"},
				{Lang: "en", Text: "Generic settings page"},
			},
		}
	}

	return thunderbirdConfig{
		Version:       "1.1",
		EmailProvider: provider,
	}
}
