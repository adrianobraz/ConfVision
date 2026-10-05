package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"confvision/src/config"
	"confvision/src/modulos/visdata"
	"confvision/src/storage"
)

func main() {
	config.ConfigurarApp()
	ctx := context.Background()
	v := strconv.FormatInt(time.Now().Unix(), 10)
	db, err := visdata.DB()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, n := range []int{1, 2, 3} {
		path := fmt.Sprintf(`C:\sistemaconfmonit\%d.jpg`, n)
		raw, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ler %s: %v\n", path, err)
			os.Exit(1)
		}
		eventoID := 990000 + n
		url, err := storage.UploadEventoSnapshot(ctx, "0", eventoID, raw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "upload %d: %v\n", eventoID, err)
			os.Exit(1)
		}
		snapURL := url + "?v=" + v
		hash := visdata.CodigoImagemPublica(eventoID)
		_, err = db.ExecContext(ctx, `
UPDATE vis_evento
SET snapshot_url = $2,
    codigo_imagem_publico = $3,
    imagem_liberada_em = NOW()
WHERE id = $1`, eventoID, snapURL, hash)
		if err != nil {
			fmt.Fprintf(os.Stderr, "postgres %d: %v\n", eventoID, err)
			os.Exit(1)
		}
		fmt.Printf("OK %d.jpg %d bytes -> evento=%d link=https://imagem.dnsid.com.br/%s\n", n, len(raw), eventoID, hash)
	}
}
