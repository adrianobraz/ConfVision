package pgreceptordns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type cfClient struct {
	token  string
	zoneID string
	http   *http.Client
}

func newCFClient(token, zoneID string) *cfClient {
	return &cfClient{
		token:  strings.TrimSpace(token),
		zoneID: strings.TrimSpace(zoneID),
		http:   &http.Client{Timeout: 25 * time.Second},
	}
}

func (c *cfClient) configured() bool {
	return c != nil && c.token != "" && c.zoneID != ""
}

type cfDNSRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
}

type cfListResp struct {
	Success bool          `json:"success"`
	Errors  []cfAPIError  `json:"errors"`
	Result  []cfDNSRecord `json:"result"`
}

type cfOneResp struct {
	Success bool         `json:"success"`
	Errors  []cfAPIError `json:"errors"`
	Result  cfDNSRecord  `json:"result"`
}

type cfAPIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *cfClient) authHeader() string {
	return "Bearer " + c.token
}

func cfErrMsg(errors []cfAPIError) string {
	if len(errors) == 0 {
		return "erro cloudflare"
	}
	return errors[0].Message
}

func (c *cfClient) recordExists(fqdn string) (bool, string, error) {
	u := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?name=%s&type=A",
		c.zoneID, url.QueryEscape(fqdn))
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return false, "", err
	}
	req.Header.Set("Authorization", c.authHeader())
	res, err := c.http.Do(req)
	if err != nil {
		return false, "", err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return false, "", fmt.Errorf("cloudflare HTTP %d: %s", res.StatusCode, string(raw))
	}
	var out cfListResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return false, "", err
	}
	if !out.Success {
		return false, "", fmt.Errorf("%s", cfErrMsg(out.Errors))
	}
	if len(out.Result) == 0 {
		return false, "", nil
	}
	return true, out.Result[0].ID, nil
}

func (c *cfClient) createA(name, ip string) (recordID string, err error) {
	body, _ := json.Marshal(map[string]any{
		"type":    "A",
		"name":    name,
		"content": ip,
		"ttl":     1,
		"proxied": false,
	})
	u := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", c.zoneID)
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return "", fmt.Errorf("cloudflare HTTP %d: %s", res.StatusCode, string(raw))
	}
	var out cfOneResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if !out.Success {
		return "", fmt.Errorf("%s", cfErrMsg(out.Errors))
	}
	return out.Result.ID, nil
}

func (c *cfClient) deleteRecord(recordID string) error {
	if strings.TrimSpace(recordID) == "" {
		return nil
	}
	u := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", c.zoneID, recordID)
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.authHeader())
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return fmt.Errorf("cloudflare HTTP %d: %s", res.StatusCode, string(raw))
	}
	var out struct {
		Success bool         `json:"success"`
		Errors  []cfAPIError `json:"errors"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return err
	}
	if !out.Success {
		return fmt.Errorf("%s", cfErrMsg(out.Errors))
	}
	return nil
}
