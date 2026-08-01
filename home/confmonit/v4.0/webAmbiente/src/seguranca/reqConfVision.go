package seguranca

import (
	"fmt"
	"io"
	"net/http"
	"webAmbiente/src/config"
)

func ReqConfVisionGET(path string) ([]byte, error) {
	if config.XanoConfVision == "" {
		return nil, fmt.Errorf("XANO_CONFVISION_API nao configurado")
	}

	url := config.XanoConfVision + path
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if config.XanoConfVisionToken != "" {
		req.Header.Set("Authorization", "Bearer "+config.XanoConfVisionToken)
	}

	res, err := xanoHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		msg := string(raw)
		if msg == "" {
			msg = res.Status
		}
		return nil, fmt.Errorf("confvision HTTP %d: %s", res.StatusCode, msg)
	}
	return raw, nil
}
