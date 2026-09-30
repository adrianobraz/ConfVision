package xano

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type MetaClient struct {
	baseURL     string
	workspaceID int
	token       string
	http        *http.Client
}

func NewMeta(baseURL, token string, workspaceID int) *MetaClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	token = strings.TrimSpace(token)
	if baseURL == "" || token == "" || workspaceID <= 0 {
		return nil
	}
	return &MetaClient{
		baseURL:     baseURL,
		workspaceID: workspaceID,
		token:       token,
		http:        &http.Client{Timeout: 90 * time.Second},
	}
}

func (c *MetaClient) Enabled() bool {
	return c != nil && c.baseURL != "" && c.token != ""
}

func (c *MetaClient) ListTableContent(tableID int) ([]map[string]any, error) {
	var all []map[string]any
	page := 1
	for {
		url := fmt.Sprintf("%s/workspace/%d/table/%d/content?page=%d&per_page=100", c.baseURL, c.workspaceID, tableID, page)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		res, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode >= 400 {
			return nil, fmt.Errorf("xano meta HTTP %d: %s", res.StatusCode, string(body))
		}
		var parsed struct {
			Items    []map[string]any `json:"items"`
			NextPage *int             `json:"nextPage"`
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, fmt.Errorf("parse meta content: %w", err)
		}
		all = append(all, parsed.Items...)
		if parsed.NextPage == nil || *parsed.NextPage <= page {
			break
		}
		page = *parsed.NextPage
	}
	return all, nil
}

func (c *MetaClient) GetRecordByID(tableID int, id int64) (map[string]any, bool, error) {
	if id <= 0 {
		return nil, false, nil
	}
	page := 1
	for {
		url := fmt.Sprintf("%s/workspace/%d/table/%d/content?page=%d&per_page=100", c.baseURL, c.workspaceID, tableID, page)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, false, err
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		res, err := c.http.Do(req)
		if err != nil {
			return nil, false, err
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode >= 400 {
			return nil, false, fmt.Errorf("xano meta HTTP %d: %s", res.StatusCode, string(body))
		}
		var parsed struct {
			Items    []map[string]any `json:"items"`
			NextPage *int             `json:"nextPage"`
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, false, fmt.Errorf("parse meta content: %w", err)
		}
		for _, item := range parsed.Items {
			if int64(intFromAny(item["id"])) == id {
				return item, true, nil
			}
		}
		if parsed.NextPage == nil || *parsed.NextPage <= page {
			break
		}
		page = *parsed.NextPage
	}
	return nil, false, nil
}

func (c *MetaClient) SearchTableContent(tableID int, payload map[string]any) ([]map[string]any, error) {
	if payload == nil {
		payload = map[string]any{}
	}
	var all []map[string]any
	page := 1
	for {
		payload["page"] = page
		payload["per_page"] = 100
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		url := fmt.Sprintf("%s/workspace/%d/table/%d/content/search", c.baseURL, c.workspaceID, tableID)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode >= 400 {
			return nil, fmt.Errorf("xano meta search HTTP %d: %s", res.StatusCode, string(body))
		}
		var parsed struct {
			Items    []map[string]any `json:"items"`
			NextPage *int             `json:"nextPage"`
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, fmt.Errorf("parse meta search: %w", err)
		}
		all = append(all, parsed.Items...)
		if parsed.NextPage == nil || *parsed.NextPage <= page {
			break
		}
		page = *parsed.NextPage
	}
	return all, nil
}
