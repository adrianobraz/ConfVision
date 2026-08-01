package storage

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"webAmbiente/src/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
	once     sync.Once
	s3Client *s3.Client
	s3Base   string
	initErr  error
)

const plantaObjectName = "planta.jpg"

func enabled() bool {
	return config.ContaboS3AccessKey != "" && config.ContaboS3SecretKey != ""
}

func initClient() {
	if !enabled() {
		initErr = fmt.Errorf("upload S3 nao configurado")
		return
	}

	endpoints := []string{
		config.ContaboS3Endpoint,
		"https://usc1.contabostorage.com",
		"https://eu2.contabostorage.com",
	}
	seen := map[string]bool{}
	var last error

	for _, ep := range endpoints {
		ep = strings.TrimRight(ep, "/")
		if !strings.HasPrefix(ep, "http") {
			ep = "https://" + ep
		}
		if seen[ep] {
			continue
		}
		seen[ep] = true

		cli := s3.New(s3.Options{
			BaseEndpoint: aws.String(ep),
			Region:       config.ContaboS3Region,
			Credentials: credentials.NewStaticCredentialsProvider(
				config.ContaboS3AccessKey,
				config.ContaboS3SecretKey,
				"",
			),
			UsePathStyle: true,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err := cli.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: aws.String(config.ContaboS3Bucket),
		})
		cancel()
		if err == nil {
			s3Client = cli
			s3Base = ep
			return
		}
		last = err
	}

	initErr = fmt.Errorf("bucket %s inacessivel: %w", config.ContaboS3Bucket, last)
}

func client() (*s3.Client, string, error) {
	once.Do(initClient)
	if initErr != nil {
		return nil, "", initErr
	}
	return s3Client, s3Base, nil
}

func mapaPrefix(mapaID int) string {
	return fmt.Sprintf("mapas/%d/", mapaID)
}

func mapaPlantaKey(mapaID int) string {
	return mapaPrefix(mapaID) + plantaObjectName
}

func PublicURL(key string) string {
	_, base, err := client()
	if err != nil {
		return ""
	}
	return buildPublicURL(base, key)
}

func buildPublicURL(base, key string) string {
	key = strings.TrimPrefix(key, "/")
	bucket := config.ContaboS3Bucket
	if config.ContaboS3TenantId != "" {
		bucket = config.ContaboS3TenantId + ":" + bucket
	}
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(base, "/"), bucket, key)
}

// NormalizeImagemURL corrige URLs antigas salvas sem o tenantId da Contabo.
func NormalizeImagemURL(url string) string {
	url = strings.TrimSpace(url)
	if url == "" || config.ContaboS3TenantId == "" {
		return url
	}
	tenantBucket := config.ContaboS3TenantId + ":" + config.ContaboS3Bucket
	if strings.Contains(url, tenantBucket+"/") {
		return url
	}
	needle := "/" + config.ContaboS3Bucket + "/"
	if idx := strings.Index(url, needle); idx >= 0 {
		suffix := url[idx+len(needle):]
		return url[:idx+1] + tenantBucket + "/" + suffix
	}
	return url
}

// IsURLPlantaMapa indica se a URL aponta para planta.jpg deste mapa no Contabo.
func IsURLPlantaMapa(mapaID int, url string) bool {
	if mapaID <= 0 || url == "" {
		return false
	}
	url = NormalizeImagemURL(url)
	return strings.Contains(url, fmt.Sprintf("mapas/%d/planta.jpg", mapaID))
}

// IsURLHospedadaContabo indica URL de imagem no bucket mapa (upload pelo sistema).
func IsURLHospedadaContabo(url string) bool {
	url = strings.TrimSpace(url)
	if url == "" {
		return false
	}
	if !strings.Contains(url, "contabostorage.com") {
		return false
	}
	return strings.Contains(url, "/mapas/")
}

// ExcluirPastaMapa remove todos os objetos em mapas/{id}/ no bucket.
func ExcluirPastaMapa(ctx context.Context, mapaID int) error {
	if mapaID <= 0 || !enabled() {
		return nil
	}
	return deletePrefix(ctx, mapaPrefix(mapaID))
}

func deletePrefix(ctx context.Context, prefix string) error {
	cli, _, err := client()
	if err != nil {
		return err
	}

	paginator := s3.NewListObjectsV2Paginator(cli, &s3.ListObjectsV2Input{
		Bucket: aws.String(config.ContaboS3Bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}
		if len(page.Contents) == 0 {
			continue
		}
		objs := make([]types.ObjectIdentifier, 0, len(page.Contents))
		for _, o := range page.Contents {
			objs = append(objs, types.ObjectIdentifier{Key: o.Key})
		}
		_, err = cli.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(config.ContaboS3Bucket),
			Delete: &types.Delete{Objects: objs},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// SubstituirImagemMapa comprime, apaga pasta do mapa e envia planta.jpg.
func SubstituirImagemMapa(ctx context.Context, mapaID int, raw []byte) (string, error) {
	if mapaID <= 0 {
		return "", fmt.Errorf("id do mapa invalido para upload")
	}
	if !enabled() {
		return "", fmt.Errorf("upload de imagem nao configurado no servidor")
	}
	if len(raw) == 0 {
		return "", fmt.Errorf("arquivo vazio")
	}

	maxUpload := config.ContaboMapaMaxUploadBytes
	if maxUpload <= 0 {
		maxUpload = 10 << 20
	}
	if len(raw) > maxUpload {
		return "", fmt.Errorf("arquivo original maior que %d MB", maxUpload/(1<<20))
	}

	compressed, err := ComprimirPlantaJPEG(raw)
	if err != nil {
		return "", err
	}

	if err := ExcluirPastaMapa(ctx, mapaID); err != nil {
		return "", err
	}

	key := mapaPlantaKey(mapaID)
	cli, base, err := client()
	if err != nil {
		return "", err
	}

	_, err = cli.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(config.ContaboS3Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(compressed),
		ContentType: aws.String("image/jpeg"),
		ACL:         types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", err
	}

	return buildPublicURL(base, key), nil
}
