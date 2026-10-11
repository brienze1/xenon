package usecase

import "strings"

func CookieDomainUrl(domain string) string {
	return "https://" + strings.TrimPrefix(strings.TrimSpace(domain), ".") + "/"
}
