package repositories

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log"
	"os"
)

type FileStorage struct {
	client   *minio.Client
	endpoint string
}

func NewFileStorage() *FileStorage {
	ACCESS_KEY, _ := os.LookupEnv("MINIO_ACCESS_KEY")
	SECRET_KEY, _ := os.LookupEnv("MINIO_SECRET_KEY")
	ENDPOINT, _ := os.LookupEnv("MINIO_ENDPOINT")

	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(ACCESS_KEY, SECRET_KEY, ""),
		Secure: false,
	}

	client, err := minio.New(ENDPOINT, opts)
	if err != nil {
		log.Fatal(err)
	}

	return &FileStorage{
		client:   client,
		endpoint: ENDPOINT,
	}
}

func (storage *FileStorage) CreateBucket(ctx context.Context, bucketName string) error {
	makeBucketOpts := minio.MakeBucketOptions{}

	err := storage.client.MakeBucket(ctx, bucketName, makeBucketOpts)

	if err != nil {
		return err
	}

	return nil
}

func (storage *FileStorage) RemoveBucket(ctx context.Context, bucketName string) error {

	err := storage.client.RemoveBucket(ctx, bucketName)

	if err != nil {
		return err
	}

	return nil
}

func (storage *FileStorage) RemoveObjects(ctx context.Context, bucketName string) error {
	listObjectOpts := minio.ListObjectsOptions{}
	removeObjectsOpts := minio.RemoveObjectsOptions{}

	objects := storage.client.ListObjects(ctx, bucketName, listObjectOpts)
	errs := storage.client.RemoveObjects(ctx, bucketName, objects, removeObjectsOpts)

	for err := range errs {
		if err.Err != nil {
			return err.Err
		}
	}

	return nil
}

func (storage *FileStorage) UploadFile(ctx context.Context, bucketName, fileName string, file io.Reader, size int64, contentType string) error {
	putObjectOpts := minio.PutObjectOptions{
		ContentType:  contentType,
		UserMetadata: map[string]string{"x-amz-acl": "public-read"},
	}

	_, err := storage.client.PutObject(ctx, bucketName, fileName, file, size, putObjectOpts)

	if err != nil {
		return err
	}

	return nil
}

func (storage *FileStorage) DownloadFile(ctx context.Context, bucketName, fileName, path string) error {
	opts := minio.GetObjectOptions{}

	_, err := storage.client.GetObject(ctx, bucketName, fileName, opts)
	if err != nil {
		return err
	}

	err = storage.client.FGetObject(ctx, bucketName, fileName, fmt.Sprintf("%s/%s", path, fileName), opts)
	if err != nil {
		return err
	}

	return nil
}

func (storage *FileStorage) DeleteFile(ctx context.Context, bucketName, fileName string) error {
	opts := minio.RemoveObjectOptions{}

	err := storage.client.RemoveObject(ctx, bucketName, fileName, opts)
	if err != nil {
		return err
	}

	return nil
}

func (storage *FileStorage) GetFile(ctx context.Context, bucketName, fileName string) (minio.ObjectInfo, error) {
	opts := minio.GetObjectOptions{}

	object, err := storage.client.GetObject(ctx, bucketName, fileName, opts)
	if err != nil {
		return minio.ObjectInfo{}, err
	}

	objectStat, err := object.Stat()
	if err != nil {
		return minio.ObjectInfo{}, err
	}

	return objectStat, nil
}

func (storage *FileStorage) GetFileList(ctx context.Context, bucketName string) []string {
	opts := minio.ListObjectsOptions{}

	list := make([]string, 0)

	for object := range storage.client.ListObjects(ctx, bucketName, opts) {
		list = append(list, object.Key)
	}

	return list
}
