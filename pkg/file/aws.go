package file

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"os"
	"path/filepath"
	"proman-backend/internal/config"
	"proman-backend/pkg/log"
	"strings"
)

var Sess *session.Session
var Uploader *s3manager.Uploader
var Downloader *s3manager.Downloader
var S3Client *s3.S3

type FilePack struct {
	Name        string
	BaseName    string
	ContentType string
	Extension   string
	Size        int64
}

type InputFilePack struct {
	FileName string
	FilePack *FilePack
}

func newInputFilePack(filename string, filepack *FilePack) *InputFilePack {
	return &InputFilePack{
		FileName: filename,
		FilePack: filepack,
	}
}

func newFilePack(f *os.File, contentType string) (*FilePack, error) {
	ff, err := os.Open(f.Name())
	if err != nil {
		return nil, err
	}
	defer func(ff *os.File) {
		err := ff.Close()
		if err != nil {
			panic(err)
		}
	}(ff)
	info, err := ff.Stat()
	if err != nil {
		return nil, err
	}
	return &FilePack{
		Name:        f.Name(),
		BaseName:    filepath.Base(f.Name()),
		ContentType: contentType,
		Extension:   strings.Split(contentType, "/")[1],
		Size:        info.Size(),
	}, nil
}

func upload(dir string, inputFilePacks *InputFilePack) (string, error) {
	locations := []string{}

	f, err := os.Open(inputFilePacks.FilePack.Name)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Error(fmt.Sprintf("Failed to close file: %s", err.Error()))
		}
	}(f)

	fileDestination := inputFilePacks.FilePack.BaseName
	if len(dir) > 0 {
		fileDestination = dir + "/" + fileDestination
	}

	uploader := s3manager.NewUploader(Sess)

	out, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(config.S3.Bucket),
		ACL:         aws.String("public-read"),
		Key:         aws.String(fileDestination),
		ContentType: &inputFilePacks.FilePack.ContentType,
		Body:        f,
	})
	if err != nil {
		log.Errorf("Failed to upload media to amazon s3 server, %v", err)
		return "", err
	}

	locations = append(locations, out.Location)
	return fileDestination, nil
}
