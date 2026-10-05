package confvision

import (
	"confvision/src/config"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ipRateLimiter struct {
	mu   sync.Mutex
	hits map[string][]time.Time
}

var imagemRateLimiter = &ipRateLimiter{hits: map[string][]time.Time{}}

func clientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func (l *ipRateLimiter) allow(ip string, perMin, perHour int) bool {
	if perMin <= 0 && perHour <= 0 {
		return true
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	prev := l.hits[ip]
	cutHour := now.Add(-time.Hour)
	cutMin := now.Add(-time.Minute)
	kept := make([]time.Time, 0, len(prev))
	countMin := 0
	countHour := 0
	for _, t := range prev {
		if t.Before(cutHour) {
			continue
		}
		kept = append(kept, t)
		countHour++
		if !t.Before(cutMin) {
			countMin++
		}
	}
	if perMin > 0 && countMin >= perMin {
		l.hits[ip] = kept
		return false
	}
	if perHour > 0 && countHour >= perHour {
		l.hits[ip] = kept
		return false
	}
	kept = append(kept, now)
	l.hits[ip] = kept
	return true
}

func imagemRateLimitAllow(r *http.Request) bool {
	return imagemRateLimiter.allow(
		clientIP(r),
		config.ImagemRateLimitPerMin,
		config.ImagemRateLimitPerHour,
	)
}
