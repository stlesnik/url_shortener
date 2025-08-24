package middleware

import (
	"github.com/stlesnik/url_shortener/internal/config"
	"net"
	"net/http"
)

func WithTrustedSubnet(cfg *config.Config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.TrustedSubnet == "" {
			http.Error(w, "IP address is not in trusted Subnet", http.StatusForbidden)
			return
		}
		_, ipNet, err := net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			http.Error(w, "Cannot parse subnet", http.StatusInternalServerError)
			return
		}

		ipStr := r.Header.Get("X-Real-IP")
		ip := net.ParseIP(ipStr)
		if ip == nil {
			http.Error(w, "Cannot find IP address", http.StatusForbidden)
			return
		}

		if ipNet.Contains(ip) {
			next(w, r)
		} else {
			http.Error(w, "IP address is not in trusted Subnet", http.StatusForbidden)
			return
		}
	}
}
