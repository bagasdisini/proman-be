package config

import (
	"os"
	"strconv"
)

var S3 struct {
	EndPoint string `mapstructure:"AWS_S3_ENDPOINT"`
	Region   string `mapstructure:"AWS_S3_REGION"`
	Bucket   string `mapstructure:"AWS_S3_BUCKET"`
}

var AWS struct {
	AccessKeyId     string `mapstructure:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	AvatarDir       string `mapstructure:"AWS_S3_AVATAR_DIR"`
	FileDir         string `mapstructure:"AWS_S3_FILE_DIR"`
	ProjectLogoDir  string `mapstructure:"AWS_S3_PROJECT_LOGO_DIR"`
}

var Upload struct {
	FileMaxSize   int64 `mapstructure:"FILE_UPLOAD_MAX_SIZE"`
	AvatarMaxSize int64 `mapstructure:"AVATAR_UPLOAD_MAX_SIZE"`
}

func initAws() {
	S3.EndPoint = os.Getenv("AWS_S3_ENDPOINT")
	S3.Region = os.Getenv("AWS_S3_REGION")
	S3.Bucket = os.Getenv("AWS_S3_BUCKET")

	if S3.EndPoint == "" {
		panic("AWS_S3_ENDPOINT is not set")
	}
	if S3.Region == "" {
		panic("AWS_S3_REGION is not set")
	}
	if S3.Bucket == "" {
		panic("AWS_S3_BUCKET is not set")
	}

	AWS.AccessKeyId = os.Getenv("AWS_ACCESS_KEY_ID")
	AWS.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	AWS.AvatarDir = os.Getenv("AWS_S3_AVATAR_DIR")
	AWS.FileDir = os.Getenv("AWS_S3_FILE_DIR")
	AWS.ProjectLogoDir = os.Getenv("AWS_S3_PROJECT_LOGO_DIR")

	if AWS.AccessKeyId == "" {
		panic("AWS_ACCESS_KEY_ID is not set")
	}
	if AWS.SecretAccessKey == "" {
		panic("AWS_SECRET_ACCESS_KEY is not set")
	}
	if AWS.AvatarDir == "" {
		panic("AWS_S3_AVATAR_DIR is not set")
	}
	if AWS.FileDir == "" {
		panic("AWS_S3_FILE_DIR is not set")
	}
	if AWS.ProjectLogoDir == "" {
		panic("AWS_S3_PROJECT_LOGO_DIR is not set")
	}

	fileMaxSize := os.Getenv("FILE_UPLOAD_MAX_SIZE")
	if fileMaxSize == "" {
		panic("FILE_UPLOAD_MAX_SIZE is not set")
	}
	intMaxSize, err := strconv.Atoi(fileMaxSize)
	if err != nil {
		panic("FILE_UPLOAD_MAX_SIZE is not valid")
	}
	Upload.FileMaxSize = int64(intMaxSize)

	avatarMaxSize := os.Getenv("AVATAR_UPLOAD_MAX_SIZE")
	if avatarMaxSize == "" {
		panic("AVATAR_UPLOAD_MAX_SIZE is not set")
	}
	intAvatarMaxSize, err := strconv.Atoi(avatarMaxSize)
	if err != nil {
		panic("AVATAR_UPLOAD_MAX_SIZE is not valid")
	}
	Upload.AvatarMaxSize = int64(intAvatarMaxSize)
}
