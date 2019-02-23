package osssync

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type OSSClient struct {
	config *OSSConfig
	client *oss.Client
	Bucket *oss.Bucket
}

func (c *OSSClient) Init(path string) error {
	bucketConfig := new(OSSConfig)
	if err := bucketConfig.Load(path); err != nil {
		return err
	}
	c.config = bucketConfig
	client, err := oss.New(
		c.config.ENDPOINT,
		c.config.ACCESS_KEY_ID,
		c.config.ACCESS_KEY_SECRET)
	if err != nil {
		return err
	}
	c.client = client
	bucket, err := client.Bucket(c.config.BUCKET_NAME)
	if err != nil {
		return err
	}
	c.Bucket = bucket
	return nil
}

func (c *OSSClient) walkDir(dirPth, suffix string) (files []string, err error) {
	files = make([]string, 0, 30)
	suffix = strings.ToUpper(suffix)
	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, filename)
		}
		return nil
	})
	return files, err
}

func (c *OSSClient) md5SumFile(file string) (clientFileMd5String string, err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	m := md5.New()
	m.Write(data)
	value := m.Sum(nil)
	clientFileMd5String = base64.StdEncoding.EncodeToString(value)
	return clientFileMd5String, nil
}

func (c *OSSClient) Upload(subPathName string) error {
	files, err := c.walkDir(subPathName, "")
	if err != nil {
		return err
	}

	for _, file := range files {
		objectKey := file

		clientFileMd5String, err := c.md5SumFile(file)
		if err != nil {
			return err
		}

		b, err := c.Bucket.IsObjectExist(objectKey)
		if err != nil {
			return err
		}

		if b {
			header, err := c.Bucket.GetObjectDetailedMeta(objectKey)
			if err != nil {
				return err
			}
			serverFileMd5String := header.Get("content-md5")

			fmt.Println(clientFileMd5String + "  " + serverFileMd5String + "  " + file)

			if clientFileMd5String == serverFileMd5String {
				fmt.Println(clientFileMd5String + "  " + file + " skipped")
				continue
			}
		}

		options := []oss.Option{
			oss.ContentMD5(clientFileMd5String),
		}
		err = c.Bucket.PutObjectFromFile(objectKey, file, options...)
		if err != nil {
			return err
		}

		fmt.Println(clientFileMd5String + "  " + file + " >>> " + objectKey + " successful")
	}
	return nil
}

func (c *OSSClient) Download(subPathName string) error {
	marker := subPathName
	for {
		lsRes, err := c.Bucket.ListObjects(oss.Marker(marker))
		if err != nil {
			return err
		}

		for _, object := range lsRes.Objects {
			filePath := path.Dir(object.Key)
			if err = os.MkdirAll(filePath, os.ModePerm); err != nil {
				fmt.Println(err.Error())
			}

			if _, err := os.Stat(object.Key); err == nil {
				clientFileMd5String, err := c.md5SumFile(object.Key)
				if err != nil {
					return err
				}

				header, err := c.Bucket.GetObjectDetailedMeta(object.Key)
				if err != nil {
					return err
				}
				serverFileMd5String := header.Get("content-md5")

				fmt.Println(clientFileMd5String + "  " + serverFileMd5String + "  " + object.Key)

				if clientFileMd5String == serverFileMd5String {
					fmt.Println(clientFileMd5String + "  " + object.Key + " skipped")
					continue
				}
			}

			if err := c.Bucket.GetObjectToFile(object.Key, object.Key); err != nil {
				return err
			}

			fmt.Println(" >>> " + object.Key + " successful")
		}

		if lsRes.IsTruncated {
			marker = lsRes.NextMarker
		} else {
			break
		}
	}
	return nil
}