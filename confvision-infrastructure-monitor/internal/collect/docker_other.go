//go:build !unix

package collect

import "net/http"

func NewDockerClient(socketPath string) *http.Client {
	return nil
}
