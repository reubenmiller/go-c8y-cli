package config

import (
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

type CliConfiguration struct {
	viper *viper.Viper
}

func NewCliConfiguration(v *viper.Viper) *CliConfiguration {
	return &CliConfiguration{
		viper: v,
	}
}

func (c *CliConfiguration) GetCookies() []*http.Cookie {
	cookies := make([]*http.Cookie, 0)
	for _, cookieValue := range c.viper.GetStringSlice("authentication.cookies") {
		parts := strings.SplitN(cookieValue, "=", 2)
		if len(parts) != 2 {
			continue
		}

		valueParts := strings.SplitN(strings.TrimSpace(parts[1]), ";", 2)

		if len(valueParts) == 0 {
			continue
		}

		cookie := &http.Cookie{
			Name:  parts[0],
			Value: valueParts[0],
			Raw:   cookieValue,
		}
		cookies = append(cookies, cookie)
	}

	return cookies
}
