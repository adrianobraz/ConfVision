package push

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2/google"
)

var (
	onceCreds sync.Once
	credsJSON []byte
	projectID string
	credsErr  error
)

func loadCreds() {
	onceCreds.Do(func() {
		path := strings.TrimSpace(os.Getenv("FIREBASE_SERVICE_ACCOUNT_FILE"))
		if path == "" {
			credsErr = fmt.Errorf("FIREBASE_SERVICE_ACCOUNT_FILE não configurado")
			return
		}
		b, err := os.ReadFile(path)
		if err != nil {
			credsErr = err
			return
		}
		credsJSON = b
		var meta struct {
			ProjectID string `json:"project_id"`
		}
		if err := json.Unmarshal(b, &meta); err != nil {
			credsErr = err
			return
		}
		projectID = meta.ProjectID
		if projectID == "" {
			credsErr = fmt.Errorf("project_id ausente no service account")
		}
	})
}

func Enabled() bool {
	loadCreds()
	return credsErr == nil && len(credsJSON) > 0 && projectID != ""
}

func accessToken(ctx context.Context) (string, error) {
	loadCreds()
	if credsErr != nil {
		return "", credsErr
	}
	creds, err := google.CredentialsFromJSON(ctx, credsJSON, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return "", err
	}
	tok, err := creds.TokenSource.Token()
	if err != nil {
		return "", err
	}
	return tok.AccessToken, nil
}

type Message struct {
	Title string
	Body  string
	Data  map[string]string
}

func SendToTokens(tokens []string, msg Message) {
	if !Enabled() {
		log.Println("[push] FCM desabilitado:", credsErr)
		return
	}
	if len(tokens) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	token, err := accessToken(ctx)
	if err != nil {
		log.Println("[push] token FCM:", err)
		return
	}

	for _, t := range tokens {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if err := sendOne(ctx, token, t, msg); err != nil {
			log.Println("[push] envio falhou:", err)
		}
	}
}

func sendOne(ctx context.Context, accessTok, deviceToken string, msg Message) error {
	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"token": deviceToken,
			"notification": map[string]string{
				"title": msg.Title,
				"body":  msg.Body,
			},
			"data": msg.Data,
			"android": map[string]interface{}{
				"priority": "high",
				"notification": map[string]interface{}{
					"channel_id": "terminal_alarmes",
					"sound":      "default",
				},
			},
			"apns": map[string]interface{}{
				"headers": map[string]string{
					"apns-priority": "10",
				},
				"payload": map[string]interface{}{
					"aps": map[string]interface{}{
						"sound": "default",
					},
				},
			},
		},
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessTok)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	respBody, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return fmt.Errorf("FCM %d: %s", res.StatusCode, string(respBody))
	}
	return nil
}
