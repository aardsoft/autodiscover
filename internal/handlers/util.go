package handlers

import (
	"fmt"
	"strings"
)

func emailLocalPart(email string) string {
	if i := strings.IndexByte(email, '@'); i > 0 {
		return email[:i]
	}
	return email
}

func emailDomain(email string) string {
	if i := strings.IndexByte(email, '@'); i >= 0 && i+1 < len(email) {
		return email[i+1:]
	}
	return ""
}

func boolOnOff(v bool) string {
	if v {
		return "on"
	}
	return "off"
}

func socketTypeToSSL(socket string) bool {
	return strings.EqualFold(socket, "SSL")
}

func socketTypeToThunderbirdSocket(socket string) string {
	switch strings.ToUpper(socket) {
	case "SSL":
		return "SSL"
	case "STARTTLS":
		return "STARTTLS"
	default:
		return "plain"
	}
}

func fullURL(scheme, host string, port int, path string) string {
	if port == 443 && scheme == "https" {
		return fmt.Sprintf("https://%s%s", host, path)
	}
	if port == 80 && scheme == "http" {
		return fmt.Sprintf("http://%s%s", host, path)
	}
	return fmt.Sprintf("%s://%s:%d%s", scheme, host, port, path)
}

func expandPlaceholders(s, email string) string {
	s = strings.ReplaceAll(s, "%EMAILADDRESS%", email)
	s = strings.ReplaceAll(s, "%EMAILLOCALPART%", emailLocalPart(email))
	s = strings.ReplaceAll(s, "%EMAILDOMAIN%", emailDomain(email))
	return s
}
