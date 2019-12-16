package minio

import (
	"github.com/minio/minio-go/v6"
	"log"
	"net/url"
	"svf-project/config"
	"time"
)

var Client *minio.Client

func Init() (*minio.Client, error) {
	conf := config.Get().Minio

	client, err := minio.New(conf.Endpoint, conf.AccessKeyID, conf.SecretAccessKey, conf.UseSSL)
	if err != nil {
		log.Fatalln(err)
	}
	Client = client

	return client, nil
}

func CreatePostUrl(objectName string) (string, map[string]string, error) {
	policy := minio.NewPostPolicy()
	policy.SetBucket(config.Get().Minio.Bucket)
	policy.SetKey(objectName)
	policy.SetExpires(time.Now().UTC().AddDate(0, 0, 1))
	//policy.SetContentType("image/jpeg")
	policy.SetContentLengthRange(1024, 1024*1024*1024) // 1KB to 1GB

	url, formData, err := Client.PresignedPostPolicy(policy)
	if err != nil {
		log.Println(err)
		return "", nil, err
	}

	return url.String(), formData, nil
}

func GetObjectUrl(objectName string) (string, error) {
	reqParams := make(url.Values)
	//reqParams.Set("response-content-disposition", "attachment; filename=\"test.jpg\"")

	// 1 week
	expiry := time.Second * 24 * 60 * 60 * 7
	presignedURL, err := Client.PresignedGetObject(config.Get().Minio.Bucket, objectName, expiry, reqParams)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return presignedURL.String(), nil
}
