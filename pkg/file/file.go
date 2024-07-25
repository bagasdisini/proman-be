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

func getFile(c echo.Context, formName string) (*FilePack, error) {
	file, err := c.FormFile(formName)
	if err != nil {
		return nil, err
	}

	headerContentType := file.Header.Get(echo.HeaderContentType)
	headerContentTypes := strings.Split(headerContentType, "/")

	if file.Size > 2_097_152*20 { // 20MB
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Maximum file size is 2MB")
	}

	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {
			log.Error(err)
		}
	}(src)

	filename := fmt.Sprintf("%v-%v-*.%v", time.Now().UnixNano(), random.String(10), headerContentTypes[len(headerContentTypes)-1])
	dst, err := os.CreateTemp(os.TempDir(), filename)
	if err != nil {
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
		return nil, err
	}

	if _, err = io.Copy(dst, src); err != nil {
		return nil, err
	}

	fp, err := newFilePack(dst, headerContentType)
	if err != nil {
		return nil, err
	}
	return fp, nil
}

func GetFileThenUpload(c echo.Context, paramName, uploadDir string) (string, error) {
	fp, err := getFile(c, paramName)
	if err != nil {
		return "", err
	}
	inputFilePack := newInputFilePack(fp.Name, fp)
	locations, err := upload(uploadDir, inputFilePack)
	if err != nil {
		return "", err
	}
	return locations, nil
}

func getFiles(c echo.Context, formName string) ([]*FilePack, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err
	}
	files := form.File[formName]

	var filePacks []*FilePack

	for _, file := range files {
		headerContentType := file.Header.Get(echo.HeaderContentType)
		headerContentTypes := strings.Split(headerContentType, "/")

		if file.Size > 2_097_152*20 { // 20MB
			return nil, echo.NewHTTPError(http.StatusBadRequest, "Maximum file size is 2MB")
		}

		src, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer func(src multipart.File) {
			err := src.Close()
			if err != nil {
				log.Error(err)
			}
		}(src)

		filename := fmt.Sprintf("%v-%v-*.%v", time.Now().UnixNano(), random.String(10), headerContentTypes[len(headerContentTypes)-1])
		dst, err := os.CreateTemp(os.TempDir(), filename)
		if err != nil {
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
			return nil, err
		}

		if _, err = io.Copy(dst, src); err != nil {
			return nil, err
		}

		fp, err := newFilePack(dst, headerContentType)
		if err != nil {
			return nil, err
		}
		filePacks = append(filePacks, fp)
	}

	return filePacks, nil
}

func GetFilesThenUpload(c echo.Context, paramName, uploadDir string) ([]string, error) {
	fps, err := getFiles(c, paramName)
	if err != nil {
		return nil, err
	}

	var locations []string
	for _, fp := range fps {
		inputFilePack := newInputFilePack(fp.Name+"."+fp.Extension, fp)
		location, err := upload(uploadDir, inputFilePack)
		if err != nil {
			return nil, err
		}
		locations = append(locations, location)
	}
	return locations, nil
}
