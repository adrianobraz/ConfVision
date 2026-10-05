package confvision

import (
	"confvision/src/config"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/speps/go-hashids/v2"
)

const (
	streamApp     = "cam"
	hashAlphabet  = "0123456789abcdefghijklmnopqrstuvwxyz"
	hashMinLength = 12
)

func hashidsCodec() (*hashids.HashID, error) {
	secret := strings.TrimSpace(config.RtmpPublishSecret)
	if secret == "" {
		return nil, fmt.Errorf("RTMP_PUBLISH_SECRET vazio")
	}
	data := hashids.NewData()
	data.Salt = secret
	data.MinLength = hashMinLength
	data.Alphabet = hashAlphabet
	return hashids.NewWithData(data)
}

// ChaveRtmp Hashids do id da câmera (min 12, alfabeto 0-9a-z), sem app.
func ChaveRtmp(cameraID int) string {
	if cameraID < 1 {
		return ""
	}
	h, err := hashidsCodec()
	if err != nil {
		return ""
	}
	s, err := h.Encode([]int{cameraID})
	if err != nil {
		return ""
	}
	return s
}

func normalizeHashToken(chave string) string {
	s := strings.Trim(strings.TrimSpace(chave), "/")
	if strings.HasPrefix(s, streamApp+"/") {
		s = s[len(streamApp)+1:]
	} else if strings.HasPrefix(s, "live/") {
		s = s[5:]
	}
	return s
}

// ParseChaveRtmp decodifica Hashids → camera id (aceita cam/{hash} ou só hash).
func ParseChaveRtmp(chave string) (int, bool) {
	s := normalizeHashToken(chave)
	if len(s) < hashMinLength {
		return 0, false
	}
	h, err := hashidsCodec()
	if err != nil {
		return 0, false
	}
	nums, err := h.DecodeWithError(s)
	if err != nil || len(nums) != 1 || nums[0] < 1 {
		return 0, false
	}
	return nums[0], true
}

// StreamPath path MediaMTX = cam/{hash}.
func StreamPath(cameraID int) string {
	chave := ChaveRtmp(cameraID)
	if chave == "" {
		return ""
	}
	return streamApp + "/" + chave
}

// MontarRtmpPublishURL URL sem query (DVR/WIFI). dvr=true adiciona / no fim.
func MontarRtmpPublishURL(base string, cameraID int, dvr bool) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	path := StreamPath(cameraID)
	if base == "" || path == "" {
		return ""
	}
	url := base + "/" + path
	if dvr {
		url += "/"
	}
	return url
}

// RtmpPublishToken legado HMAC url-safe do id da câmera.
func RtmpPublishToken(cameraID int) string {
	secret := strings.TrimSpace(config.RtmpPublishSecret)
	if secret == "" || cameraID < 1 {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strconv.Itoa(cameraID)))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
