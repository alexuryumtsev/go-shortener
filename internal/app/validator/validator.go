// Package validator содержит функции для валидации входных данных.
package validator

import (
	"fmt"
	"net/url"
	"regexp"
)

// ValidateServerAddress проверяет формат host:port.
func ValidateServerAddress(addr string) error {
	hostPortPattern := `^([a-zA-Z0-9.-]+)?(:[0-9]+)$`
	matched, err := regexp.MatchString(hostPortPattern, addr)
	if err != nil || !matched {
		return fmt.Errorf("invalid server address format, expected host:port")
	}
	return nil
}

// ValidateBaseURL проверяет корректность URL.
func ValidateBaseURL(baseURL string) error {
	_, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %v", err)
	}
	return nil
}
