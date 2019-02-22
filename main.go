package main

import (
	"crypto/md5"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/micro/go-config"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
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

func md5SumFile(file string) (value []byte, err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	m := md5.New()
	m.Write(data)
	value = m.Sum(nil)
	return value, nil
}

func main() {

	sub_path_name := flag.String("n", "", "sub path name")
	mode := flag.String("m", "", "upload or download")
	flag.Parse()

	if *mode != "up" && *mode != "down" && *mode != "upload" && *mode != "download" {
		fmt.Println("mode is necessary! use -m up or -m down")
		return
	}

	err := config.LoadFile("./sync_config.yaml")
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

	if *mode == "up" || *mode == "upload" {
		files, err := walk_dir("./vendor", "")
		if err != nil {
			fmt.Println(err.Error())
		}

		var objectKey = ""
		for _, file := range files {
			if *sub_path_name != "" {
				objectKey = *sub_path_name + "/" + file
			}

			fileMd5, err := md5SumFile(file)
			if err != nil {
				fmt.Println(err.Error())
			}
			clientFileMd5String := base64.StdEncoding.EncodeToString(fileMd5)

			b, err := bucket.IsObjectExist(objectKey)
			if err != nil {
				fmt.Println(err.Error())
			}

			if b {
				h, err := bucket.GetObjectDetailedMeta(objectKey)
				if err != nil {
					fmt.Println(err.Error())
				}
				serverFileMd5String := h.Get("content-md5")

				fmt.Println(clientFileMd5String + "  " + serverFileMd5String + "  " + file)

				if clientFileMd5String == serverFileMd5String {
					fmt.Println(clientFileMd5String + "  " + file + " skipped")
					continue
				}
			}

			options := []oss.Option{
				oss.ContentMD5(clientFileMd5String),
			}
			err = bucket.PutObjectFromFile(objectKey, file, options...)
			if err != nil {
				fmt.Println(err.Error())
			}

			fmt.Println(clientFileMd5String + "  " + file + " >>> " + objectKey + " successful")
		}
	}

	if *mode == "down" || *mode == "download" {
		marker := *sub_path_name
		for {
			lsRes, err := bucket.ListObjects(oss.Marker(marker))
			if err != nil {
				fmt.Println(err.Error())
			}

			for _, object := range lsRes.Objects {
				filePath := path.Dir(object.Key)
				err = os.MkdirAll(filePath, os.ModePerm)
				if err != nil {
					fmt.Println(err.Error())
				}

				if _, err := os.Stat(object.Key); err == nil {
					fileMd5, err := md5SumFile(object.Key)
					if err != nil {
						fmt.Println(err.Error())
					}
					clientFileMd5String := base64.StdEncoding.EncodeToString(fileMd5)

					h, err := bucket.GetObjectDetailedMeta(object.Key)
					if err != nil {
						fmt.Println(err.Error())
					}
					serverFileMd5String := h.Get("content-md5")

					fmt.Println(clientFileMd5String + "  " + serverFileMd5String + "  " + object.Key)

					if clientFileMd5String == serverFileMd5String {
						fmt.Println(clientFileMd5String + "  " + object.Key + " skipped")
						continue
					}
				}

				err = bucket.GetObjectToFile(object.Key, object.Key)
				if err != nil {
					fmt.Println(err.Error())
				}

				fmt.Println(" >>> " + object.Key + " successful")
			}

			if lsRes.IsTruncated {
				marker = lsRes.NextMarker
			} else {
				break
			}
		}
	}
}
