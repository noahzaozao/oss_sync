package main

import (
	"github.com/micro/go-config"
)

type BucketConfig struct {
	ENDPOINT          string `json:"ENDPOINT"`
	ACCESS_KEY_ID     string `json:"ACCESS_KEY_ID"`
	ACCESS_KEY_SECRET string `json:"ACCESS_KEY_SECRET"`
	BUCKET_NAME       string `json:"BUCKET_NAME"`
}

func (bc *BucketConfig) Load(path string) error {
	if err := config.LoadFile("./sync_config.yaml"); err != nil {
		return err
	}
	if err := config.Get("config").Scan(bc); err != nil {
		return err
	}
	return nil
}
