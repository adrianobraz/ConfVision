package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// CapturarFrameRTSP captura um frame JPEG via ffmpeg a partir do stream RTSP.
func CapturarFrameRTSP(ctx context.Context, rtspURL string) ([]byte, error) {
	if rtspURL == "" {
		return nil, fmt.Errorf("url RTSP vazia")
	}

	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	tmp, err := os.CreateTemp("", "cv-snap-*.jpg")
	if err != nil {
		return nil, err
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	args := []string{
		"-hide_banner", "-loglevel", "error",
		"-rtsp_transport", "tcp",
		"-i", rtspURL,
		"-frames:v", "1",
		"-update", "1",
		"-y", tmpPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		log.Printf("confvision snapshot ffmpeg falhou (%s): %s", rtspURL, msg)
		return nil, fmt.Errorf("ffmpeg: %s", msg)
	}

	raw, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("ffmpeg nao retornou imagem; verifique se a camera esta transmitindo (path Hashids)")
	}

	return raw, nil
}
