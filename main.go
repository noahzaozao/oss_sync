package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/micro/go-config"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"crypto/md5"
)

func walk_dir(dirPth, suffix string) (files []string, err error) {
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


type BucketConfig struct {
	ENDPOINT string `json:"ENDPOINT"`
	ACCESS_KEY_ID string `json:"ACCESS_KEY_ID"`
	ACCESS_KEY_SECRET string `json:"ACCESS_KEY_SECRET"`
	BUCKET_NAME string `json:"BUCKET_NAME"`
}

func md5SumFile(file string) (value [md5.Size]byte, err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	value = md5.Sum(data)
	return value, nil
}

func main() {

	name := flag.String("n", "", "micro service name")
	flag.Parse()

	files, err := walk_dir("./vendor", "")
	if err != nil {
		fmt.Println(err.Error())
	}

	err = config.LoadFile("./config.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}

	var bucketConfig BucketConfig
	err = config.Get("config").Scan(&bucketConfig)
	if err != nil {
		fmt.Println(err.Error())
	}

	client, err := oss.New(
		bucketConfig.ENDPOINT,
		bucketConfig.ACCESS_KEY_ID,
		bucketConfig.ACCESS_KEY_SECRET)
	if err != nil {
		fmt.Println(err.Error())
	}

	//_, err = client.Bucket(bucketConfig.BUCKET_NAME)
	bucket, err := client.Bucket(bucketConfig.BUCKET_NAME)
	if err != nil {
		fmt.Println(err.Error())
	}

	var objectKey = ""
	for _, file := range files {
		if *name != "" {
			objectKey = *name + "/" + file
		}

		fileMd5, err := md5SumFile(file)
		if err != nil {
			fmt.Println(err.Error())
		}
		clientFileMd5String := hex.EncodeToString(fileMd5[:])

		b, err := bucket.IsObjectExist(objectKey)
		if err != nil {
			fmt.Println(err.Error())
		}

		if b {
			h, err := bucket.GetObjectDetailedMeta(objectKey)
			if err != nil {
				fmt.Println(err.Error())
			}
			serverFileMd5String := h.Get("Content-MD5")

			if clientFileMd5String == serverFileMd5String {
				fmt.Println(clientFileMd5String + "  " + file + " skipped")
				continue
			}
		}

		options := []oss.Option{
			oss.Meta("Content-MD5", clientFileMd5String),
		}
		err = bucket.PutObjectFromFile(objectKey, file, options...)
		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Println(clientFileMd5String + "  " + file + " >>> " + objectKey + " successful")
	}

}
