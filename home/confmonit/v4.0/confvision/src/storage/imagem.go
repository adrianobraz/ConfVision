package storage

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"confvision/src/config"

	"github.com/disintegration/imaging"
)

func ComprimirCadSnapshotJPEG(data []byte) ([]byte, error) {
	maxBytes := config.CadSnapshotMaxBytes
	if maxBytes <= 0 {
		maxBytes = 51200
	}
	maxWidth := config.CadSnapshotMaxWidth
	if maxWidth <= 0 {
		maxWidth = 1280
	}

	img, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("imagem invalida ou formato nao suportado")
	}

	w := img.Bounds().Dx()
	width := w
	if width > maxWidth {
		width = maxWidth
	}

	for width >= 320 {
		resized := img
		if w > width {
			resized = imaging.Resize(img, width, 0, imaging.Lanczos)
		}
		for q := 85; q >= 35; q -= 5 {
			var buf bytes.Buffer
			if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: q}); err != nil {
				return nil, err
			}
			if buf.Len() <= maxBytes {
				return buf.Bytes(), nil
			}
		}
		width = width * 3 / 4
	}

	return nil, fmt.Errorf("nao foi possivel reduzir a imagem para %d KB", maxBytes/1024)
}
