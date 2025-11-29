package mini

import (
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	MinClient *minio.Client
}

const DocumentBucket = "file-buc"

func MinioConnect(endpoint, accessKey, secretAccesssKey string, useSSL bool) (*MinioClient, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretAccesssKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	log.Println("connection to Minio is successful")
	min := &MinioClient{MinClient: minioClient}

	if err := min.EnsureBucket(context.Background(), DocumentBucket); err != nil {
		log.Printf("failed to create bucket: %s: %v", DocumentBucket, err)
		return nil, err
	}

	return min, nil
}

func (m *MinioClient) EnsureBucket(ctx context.Context, bucket string) error {
	exists, err := m.MinClient.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		return m.MinClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	}
	return nil
}
