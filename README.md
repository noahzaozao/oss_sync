# oss_sync

## Installation

go get -v -u github.com/noahzaozao/oss_sync

## Usage

#### Upload

oss_sync -n sub_path_name/sub_path -m up

#### Download

oss_sync -n sub_path_name/sub_path -m down

### sync_config.yaml

config:

> ENDPOINT: "http://oss-cn-shanghai-internal.aliyuncs.com"

> ACCESS_KEY_ID: ""

> ACCESS_KEY_SECRET: ""

> BUCKET_NAME: ""