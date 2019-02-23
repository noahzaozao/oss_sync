package main

import (
	"flag"
	"fmt"
	"github.com/noahzaozao/oss_sync/osssync"
	"os"
)

func main() {

	subPathName := flag.String("n", "", "sub path name")
	mode := flag.String("m", "", "upload or download")

	flag.Parse()

	if *subPathName == "" {
		fmt.Println("name is necessary! use -n path_name")
		return
	}

	if _, err := os.Stat("./sync_config.yaml"); err != nil {
		fmt.Println(err.Error())
		return
	}

	if *mode != "up" && *mode != "down" && *mode != "upload" && *mode != "download" {
		fmt.Println("mode is necessary! use -m up or -m down")
		return
	}

	ossClient := new(osssync.OSSClient)
	if err := ossClient.Init("./sync_config.yaml"); err != nil {
		fmt.Println(err.Error())
	}

	if *mode == "up" || *mode == "upload" {
		if err := ossClient.Upload(*subPathName); err != nil {
			fmt.Println(err.Error())
		}
	}

	if *mode == "down" || *mode == "download" {

		if err := ossClient.Download(*subPathName); err != nil {
			fmt.Println(err.Error())
		}

	}
}
