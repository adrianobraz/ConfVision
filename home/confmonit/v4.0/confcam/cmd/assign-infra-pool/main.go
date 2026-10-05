// assign-infra-pool — preenche vis_mediamtx_node_id e worker_id a partir do pool do franqueado.
// Uso: go run ./cmd/assign-infra-pool -camera_id=33
// Requer .env ConfCam (POSTGRES_URL) e migration 022 aplicada.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"confvision/src/config"
	"confvision/src/modulos/visdata"

	"github.com/joho/godotenv"
)

func main() {
	cameraID := flag.Int("camera_id", 0, "vis_camera.id")
	idFra := flag.String("id_franqueado", "", "atribuir todas as cameras deste franqueado sem worker ou sem nó")
	dry := flag.Bool("dry_run", false, "apenas listar")
	flag.Parse()

	_ = godotenv.Load()
	config.ConfigurarApp()
	if !config.VisPostgresEnabled {
		log.Fatal("VIS_POSTGRES_ENABLED necessario")
	}

	ctx := context.Background()
	if *cameraID > 0 {
		if err := assignOne(ctx, *cameraID, *dry); err != nil {
			log.Fatal(err)
		}
		return
	}
	if strings.TrimSpace(*idFra) != "" {
		if err := assignFranqueado(ctx, strings.TrimSpace(*idFra), *dry); err != nil {
			log.Fatal(err)
		}
		return
	}
	fmt.Fprintln(os.Stderr, "informe -camera_id ou -id_franqueado")
	os.Exit(2)
}

func assignOne(ctx context.Context, cameraID int, dry bool) error {
	cam, err := visdata.GetCameraByID(ctx, cameraID)
	if err != nil {
		return err
	}
	idFra := strings.TrimSpace(fmt.Sprint(cam["id_franqueado"]))
	if err := visdata.ValidateFranqueadoInfraPool(ctx, idFra); err != nil {
		return err
	}
	out, err := visdata.AssignCameraInfraFromPool(ctx, cameraID, dry)
	if err != nil {
		return err
	}
	fmt.Printf("camera %d: %+v\n", cameraID, out)
	return nil
}

func assignFranqueado(ctx context.Context, idFra string, dry bool) error {
	if err := visdata.ValidateFranqueadoInfraPool(ctx, idFra); err != nil {
		return err
	}
	ids, err := visdata.ListCameraIDsNeedingInfraAssign(ctx, idFra)
	if err != nil {
		return err
	}
	for _, cid := range ids {
		out, err := visdata.AssignCameraInfraFromPool(ctx, cid, dry)
		if err != nil {
			log.Printf("camera %d: %v", cid, err)
			continue
		}
		fmt.Printf("camera %d: %+v\n", cid, out)
	}
	return nil
}
