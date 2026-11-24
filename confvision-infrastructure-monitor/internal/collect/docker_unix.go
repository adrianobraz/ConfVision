//go:build unix

package collect

import (
	"context"
	"net"
	"net/http"
	"time"
)

func NewDockerClient(socketPath string) *http.Client {
	return &http.Client{
		Timeout: 4 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, "unix", socketPath)
			},
		},
	}
}
