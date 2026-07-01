package handlers

import (
	"encoding/xml"
	"io"
	"net/http"
	"strings"

	"github.com/aardsoft/autodiscover/internal/config"
)

// MS Autodiscover request namespaces
const (
	msAutodiscoverNS       = "http://schemas.microsoft.com/exchange/autodiscover/responseschema/2006"
	msOutlookResponseNS    = "http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a"
	msMobileSyncResponseNS = "http://schemas.microsoft.com/exchange/autodiscover/mobilesync/responseschema/2006"
	msRequestNS            = "http://schemas.microsoft.com/exchange/autodiscover/outlook/requestschema/2006"
	msMobileSyncRequestNS  = "http://schemas.microsoft.com/exchange/autodiscover/mobilesync/requestschema/2006"
)

type msAutodiscoverRequest struct {
	XMLName xml.Name       `xml:"Autodiscover"`
	Request msRequestInner `xml:"Request"`
}

type msRequestInner struct {
	EMailAddress             string `xml:"EMailAddress"`
	AcceptableResponseSchema string `xml:"AcceptableResponseSchema"`
}

// Outlook response structures

type msOutlookResponse struct {
	XMLName  xml.Name                 `xml:"Autodiscover"`
	XMLNS    string                   `xml:"xmlns,attr"`
	Response msOutlookAccountResponse `xml:"Response"`
}

type msOutlookAccountResponse struct {
	XMLNS   string           `xml:"xmlns,attr"`
	User    *msUser          `xml:"User,omitempty"`
	Account msOutlookAccount `xml:"Account"`
}

type msUser struct {
	DisplayName string `xml:"DisplayName"`
}

type msOutlookAccount struct {
	AccountType string       `xml:"AccountType"`
	Action      string       `xml:"Action"`
	ServiceHome string       `xml:"ServiceHome,omitempty"`
	Protocols   []msProtocol `xml:"Protocol"`
}

type msProtocol struct {
	Type           string `xml:"Type"`
	TTL            int    `xml:"TTL,omitempty"`
	Server         string `xml:"Server,omitempty"`
	Port           int    `xml:"Port,omitempty"`
	DomainRequired string `xml:"DomainRequired,omitempty"`
	SPA            string `xml:"SPA,omitempty"`
	SSL            string `xml:"SSL,omitempty"`
	AuthRequired   string `xml:"AuthRequired,omitempty"`
	UsePOPAuth     string `xml:"UsePOPAuth,omitempty"`
	SMTPLast       string `xml:"SMTPLast,omitempty"`
	LoginName      string `xml:"LoginName,omitempty"`
	// Exchange fields
	ASUrl  string `xml:"ASUrl,omitempty"`
	OOFUrl string `xml:"OOFUrl,omitempty"`
	OABUrl string `xml:"OABUrl,omitempty"`
	EwsUrl string `xml:"EwsUrl,omitempty"`
}

// MobileSync response structures

type msMobileSyncResponse struct {
	XMLName  xml.Name                  `xml:"Autodiscover"`
	XMLNS    string                    `xml:"xmlns,attr"`
	Response msMobileSyncInnerResponse `xml:"Response"`
}

type msMobileSyncInnerResponse struct {
	XMLNS   string             `xml:"xmlns,attr"`
	Culture string             `xml:"Culture"`
	User    msMobileSyncUser   `xml:"User"`
	Action  msMobileSyncAction `xml:"Action"`
}

type msMobileSyncUser struct {
	DisplayName  string `xml:"DisplayName"`
	EMailAddress string `xml:"EMailAddress"`
}

type msMobileSyncAction struct {
	Settings msMobileSyncSettings `xml:"Settings"`
}

type msMobileSyncSettings struct {
	Server msMobileSyncServer `xml:"Server"`
}

type msMobileSyncServer struct {
	Type string `xml:"Type"`
	Url  string `xml:"Url"`
	Name string `xml:"Name"`
}

// HandleMSAutodiscover processes POST requests to the Microsoft Autodiscover endpoint.
func HandleMSAutodiscover(s *config.Settings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var req msAutodiscoverRequest
		if err := xml.Unmarshal(body, &req); err != nil {
			// Try to serve a generic response if we can't parse
			serveGenericOutlookResponse(w, s, "")
			return
		}

		email := strings.TrimSpace(req.Request.EMailAddress)
		schema := strings.TrimSpace(req.Request.AcceptableResponseSchema)

		switch {
		case strings.HasSuffix(schema, "/mobilesync/responseschema/2006"):
			if s.ActiveSync == nil || s.ActiveSync.URL == "" {
				http.Error(w, "ActiveSync not configured", http.StatusNotFound)
				return
			}
			resp := msMobileSyncResponse{
				XMLNS: msAutodiscoverNS,
				Response: msMobileSyncInnerResponse{
					XMLNS:   msMobileSyncResponseNS,
					Culture: "en:us",
					User: msMobileSyncUser{
						DisplayName:  emailLocalPart(email),
						EMailAddress: email,
					},
					Action: msMobileSyncAction{
						Settings: msMobileSyncSettings{
							Server: msMobileSyncServer{
								Type: "MobileSync",
								Url:  s.ActiveSync.URL,
								Name: s.ActiveSync.URL,
							},
						},
					},
				},
			}
			writeXML(w, http.StatusOK, resp)
		default:
			serveOutlookResponse(w, s, email)
		}
	}
}

// HandleMSAutodiscoverGET serves Outlook settings for GET requests.
func HandleMSAutodiscoverGET(s *config.Settings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := strings.TrimSpace(r.URL.Query().Get("email"))
		serveOutlookResponse(w, s, email)
	}
}

func serveOutlookResponse(w http.ResponseWriter, s *config.Settings, email string) {
	resp := msOutlookResponse{
		XMLNS: msAutodiscoverNS,
		Response: msOutlookAccountResponse{
			XMLNS: msOutlookResponseNS,
			User: &msUser{
				DisplayName: s.Info.Name,
			},
			Account: msOutlookAccount{
				AccountType: "email",
				Action:      "settings",
				ServiceHome: s.Info.URL,
				Protocols:   buildOutlookProtocols(s, email),
			},
		},
	}
	writeXML(w, http.StatusOK, resp)
}

func serveGenericOutlookResponse(w http.ResponseWriter, s *config.Settings, email string) {
	serveOutlookResponse(w, s, email)
}

func buildOutlookProtocols(s *config.Settings, email string) []msProtocol {
	var protos []msProtocol

	// IMAP protocol
	if s.Server.IMAP.Host != "" {
		imap := msProtocol{
			Type:           "IMAP",
			TTL:            s.TTL,
			Server:         s.Server.IMAP.Host,
			Port:           s.Server.IMAP.Port,
			DomainRequired: boolOnOff(s.Server.DomainRequired),
			SPA:            "off",
			SSL:            boolOnOff(socketTypeToSSL(s.Server.IMAP.Socket)),
			AuthRequired:   "on",
			UsePOPAuth:     "on",
			SMTPLast:       "off",
		}
		if email != "" {
			imap.LoginName = email
		}
		protos = append(protos, imap)
	}

	// SMTP protocol
	if s.Server.SMTP.Host != "" {
		protos = append(protos, msProtocol{
			Type:           "SMTP",
			TTL:            s.TTL,
			Server:         s.Server.SMTP.Host,
			Port:           s.Server.SMTP.Port,
			DomainRequired: boolOnOff(s.Server.DomainRequired),
			SPA:            "off",
			SSL:            boolOnOff(socketTypeToSSL(s.Server.SMTP.Socket)),
			AuthRequired:   "on",
			UsePOPAuth:     "on",
			SMTPLast:       "off",
		})
	}

	// Exchange protocols
	if s.Exchange != nil {
		if s.Exchange.EWSURL != "" {
			// EXPR = external Outlook Anywhere / MAPI over HTTP
			protos = append(protos, msProtocol{
				Type:   "EXPR",
				ASUrl:  s.Exchange.EWSURL,
				OOFUrl: s.Exchange.OOFURL,
				OABUrl: s.Exchange.OABURL,
				EwsUrl: s.Exchange.EWSURL,
			})
			// EXCH = internal
			protos = append(protos, msProtocol{
				Type:   "EXCH",
				ASUrl:  s.Exchange.EWSURL,
				OOFUrl: s.Exchange.OOFURL,
				OABUrl: s.Exchange.OABURL,
				EwsUrl: s.Exchange.EWSURL,
			})
		}
		// WEB = Outlook Web Access
		if s.Info.URL != "" {
			protos = append(protos, msProtocol{
				Type:   "WEB",
				Server: s.Info.URL,
			})
		}
	}

	return protos
}
