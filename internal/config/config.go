package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Settings holds the runtime configuration loaded from settings.json.
type Settings struct {
	Info struct {
		Name   string `json:"name"`
		URL    string `json:"url"`
		Domain string `json:"domain"`
	} `json:"info"`
	Server struct {
		IMAP struct {
			Host   string `json:"host"`
			Port   int    `json:"port"`
			Socket string `json:"socket"`
		} `json:"imap"`
		SMTP struct {
			Host   string `json:"host"`
			Port   int    `json:"port"`
			Socket string `json:"socket"`
		} `json:"smtp"`
		DomainRequired bool `json:"domain_required"`
	} `json:"server"`
	TTL        int                 `json:"ttl"`
	CalDAV     *CalDAVSettings     `json:"caldav,omitempty"`
	CardDAV    *CardDAVSettings    `json:"carddav,omitempty"`
	Exchange   *ExchangeSettings   `json:"exchange,omitempty"`
	ActiveSync *ActiveSyncSettings `json:"activesync,omitempty"`
	VPN        *VPNSettings        `json:"vpn,omitempty"`
	MDM        *MDMSettings        `json:"mdm,omitempty"`
}

// CalDAVSettings describes a CalDAV server endpoint.
type CalDAVSettings struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	SSL          bool   `json:"ssl"`
	PrincipalURL string `json:"principal_url"`
}

// CardDAVSettings describes a CardDAV server endpoint.
type CardDAVSettings struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	SSL          bool   `json:"ssl"`
	PrincipalURL string `json:"principal_url"`
}

// ExchangeSettings holds Microsoft Exchange Web Services URLs.
type ExchangeSettings struct {
	EWSURL string `json:"ews_url,omitempty"`
	OABURL string `json:"oab_url,omitempty"`
	OOFURL string `json:"oof_url,omitempty"`
}

// ActiveSyncSettings holds the Exchange ActiveSync endpoint URL.
type ActiveSyncSettings struct {
	URL string `json:"url"`
}

// VPNSettings describes a VPN configuration for Apple profiles.
type VPNSettings struct {
	Type       string `json:"type"`
	Host       string `json:"host"`
	RemoteID   string `json:"remote_id,omitempty"`
	LocalID    string `json:"local_id,omitempty"`
	AuthMethod string `json:"auth_method,omitempty"`
}

// MDMSettings holds Mobile Device Management enrollment information.
type MDMSettings struct {
	ServerURL string `json:"server_url"`
}

// LoadSettings reads and parses the JSON configuration at path.
func LoadSettings(path string) (*Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading settings: %w", err)
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing settings: %w", err)
	}
	if s.TTL == 0 {
		s.TTL = 1
	}
	return &s, nil
}
