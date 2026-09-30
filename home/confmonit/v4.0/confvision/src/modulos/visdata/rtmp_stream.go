package visdata

import (
	"confvision/src/config"
	"context"
	"fmt"
	"strings"

	"github.com/speps/go-hashids/v2"
)

const streamAppCam = "cam"

func rtmpHashCodec() (*hashids.HashID, error) {
	secret := strings.TrimSpace(config.RtmpPublishSecret)
	if secret == "" {
		return nil, fmt.Errorf("RTMP_PUBLISH_SECRET vazio")
	}
	data := hashids.NewData()
	data.Salt = secret
	data.MinLength = 12
	data.Alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	return hashids.NewWithData(data)
}

// CameraStreamPath retorna cam/{hash12} para HLS/RTMP (MediaMTX).
func CameraStreamPath(cameraID int) string {
	if cameraID < 1 {
		return ""
	}
	h, err := rtmpHashCodec()
	if err != nil {
		return ""
	}
	s, err := h.Encode([]int{cameraID})
	if err != nil || s == "" {
		return ""
	}
	return streamAppCam + "/" + s
}

// EnrichCameraStreamURLs preenche stream_path e hls_url no mapa da camera (terminal/web).
func EnrichCameraStreamURLs(ctx context.Context, cam map[string]any) {
	if cam == nil {
		return
	}
	id := intVal(cam, "id")
	if id <= 0 {
		return
	}
	path := CameraStreamPath(id)
	if path == "" {
		return
	}
	cam["stream_path"] = path

	hlsBase := strings.TrimRight(strings.TrimSpace(config.MediamtxHlsPublic), "/")
	if hlsBase == "" {
		hlsBase = strings.TrimRight(strings.TrimSpace(config.MediamtxHlsBase), "/")
	}
	if mtx, err := ResolveMediamtxForCamera(ctx, id, false); err == nil {
		if pub := strings.TrimSpace(fmt.Sprint(mtx["hls_public"])); pub != "" && pub != "<nil>" {
			hlsBase = strings.TrimRight(pub, "/")
		}
	}
	if hlsBase != "" {
		cam["hls_url"] = hlsBase + "/" + path + "/index.m3u8"
	}
}
