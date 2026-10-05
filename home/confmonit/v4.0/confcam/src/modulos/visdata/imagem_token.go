package visdata

import (
	"confvision/src/config"
	"fmt"
	"strings"

	"github.com/speps/go-hashids/v2"
)

const (
	imagemHashAlphabet  = "0123456789abcdefghijklmnopqrstuvwxyz"
	imagemHashMinLength = 12
)

func imagemHashCodec() (*hashids.HashID, error) {
	secret := strings.TrimSpace(config.ImagemPublicSecret)
	if secret == "" {
		return nil, fmt.Errorf("IMAGEM_PUBLIC_SECRET vazio")
	}
	data := hashids.NewData()
	data.Salt = secret
	data.MinLength = imagemHashMinLength
	data.Alphabet = imagemHashAlphabet
	return hashids.NewWithData(data)
}

// CodigoImagemPublica gera Hashids opaco do vis_evento_id (min 12 chars).
func CodigoImagemPublica(eventoID int) string {
	if eventoID < 1 {
		return ""
	}
	h, err := imagemHashCodec()
	if err != nil {
		return ""
	}
	s, err := h.Encode([]int{eventoID})
	if err != nil {
		return ""
	}
	return s
}

// ParseCodigoImagemPublica decodifica codigo → vis_evento_id.
func ParseCodigoImagemPublica(codigo string) (int, bool) {
	s := strings.Trim(strings.TrimSpace(codigo), "/")
	if len(s) < imagemHashMinLength {
		return 0, false
	}
	h, err := imagemHashCodec()
	if err != nil {
		return 0, false
	}
	nums, err := h.DecodeWithError(s)
	if err != nil || len(nums) != 1 || nums[0] < 1 {
		return 0, false
	}
	return nums[0], true
}

// ComplementoImagemMoni formato 046|1|1|{hash}?
func ComplementoImagemMoni(codigoIntegracao, codigoImagem string) string {
	prefixo := strings.TrimSpace(codigoIntegracao)
	if prefixo == "" {
		prefixo = "046"
	}
	codigo := strings.TrimSpace(codigoImagem)
	if codigo == "" {
		return ""
	}
	return fmt.Sprintf("%s|1|1|%s?", prefixo, codigo)
}
