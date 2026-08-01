package storage

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"confvision/src/config"

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

func cadSnapshotKey(cameraID int) string {
	return fmt.Sprintf("confvision/cad_snapshot/%d.jpg", cameraID)
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

func deleteObject(ctx context.Context, key string) error {
	cli, _, err := client()
	if err != nil {
		return err
	}
	_, err = cli.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(config.ContaboS3Bucket),
		Key:    aws.String(key),
	})
	return err
}

// SubstituirCadSnapshot comprime JPEG, remove arquivo anterior e envia novo snapshot.
func SubstituirCadSnapshot(ctx context.Context, cameraID int, raw []byte) (string, error) {
	if cameraID <= 0 {
		return "", fmt.Errorf("id da camera invalido para upload")
	}
	if !enabled() {
		return "", fmt.Errorf("upload de imagem nao configurado no servidor")
	}
	if len(raw) == 0 {
		return "", fmt.Errorf("arquivo vazio")
	}

	compressed, err := ComprimirCadSnapshotJPEG(raw)
	if err != nil {
		return "", err
	}

	key := cadSnapshotKey(cameraID)
	if err := deleteObject(ctx, key); err != nil {
		return "", err
	}

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

// RemoverCadSnapshot apaga o JPEG de cadastro no object storage (ignora se nao configurado).
func RemoverCadSnapshot(ctx context.Context, cameraID int) error {
	if cameraID <= 0 {
		return fmt.Errorf("id da camera invalido")
	}
	if !enabled() {
		return nil
	}
	return deleteObject(ctx, cadSnapshotKey(cameraID))
}
