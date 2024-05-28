package file

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"proman-backend/pkg/log"
	"strings"
	"time"
)

type Image struct {
	Name string
	Type string

	FileName string
}

func GetFileImage(c echo.Context, formName string) (*Image, error) {
	file, err := c.FormFile(formName)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	headerContentType := file.Header.Get(echo.HeaderContentType)
	if headerContentType != "image/jpeg" && headerContentType != "image/png" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "The file format is not supported (JPEG and PNG only)")
	}

	if file.Size > 2_097_152 { // 2MB
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Maximum file size is 2MB")
	}

	src, err := file.Open()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {
			log.Error(err)
		}
	}(src)

	filename := fmt.Sprintf("*-%v-%v.%v", random.String(10), time.Now().UnixNano(), strings.TrimPrefix(headerContentType, "image/"))
	dst, err := os.CreateTemp(os.TempDir(), filename)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer func(dst *os.File) {
		err := dst.Close()
		if err != nil {
			log.Error(err)
		}
	}(dst)

	_, err = src.Seek(0, io.SeekStart)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if _, err = io.Copy(dst, src); err != nil {
		log.Error(err)
		return nil, err
	}

	return &Image{
		Name:     file.Filename,
		Type:     strings.Split(headerContentType, "/")[1],
		FileName: dst.Name(),
	}, nil
}
