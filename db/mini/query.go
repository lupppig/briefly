package mini

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/minio/minio-go/v7"
)

func (m *MinioClient) AddToBucket(ctx context.Context, file multipart.File, header *multipart.FileHeader, bucket, objectPath string) (string, error) {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, file)
	if err != nil {
		return "", fmt.Errorf("failed to read file into buffer: %w", err)
	}

	_, err = m.MinClient.PutObject(
		ctx,
		bucket,
		objectPath,
		bytes.NewReader(buf.Bytes()),
		int64(buf.Len()),
		minio.PutObjectOptions{
			ContentType: header.Header.Get("Content-Type"),
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to bucket: %w", err)
	}

	return objectPath, nil
}

func (m *MinioClient) GetObjectBuffer(bucket, objectKey string) (*bytes.Buffer, error) {
	obj, err := m.MinClient.GetObject(context.Background(), bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer obj.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, obj); err != nil {
		return nil, fmt.Errorf("failed to copy object content to buffer: %w", err)
	}
	return buf, nil
}

func (m *MinioClient) ObjectExists(bucket, objectKey string) (bool, error) {
	obj, err := m.MinClient.GetObject(context.Background(), bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get object handle: %w", err)
	}
	defer obj.Close()

	_, err = obj.Stat()
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to stat object: %w", err)
	}

	return true, nil
}
