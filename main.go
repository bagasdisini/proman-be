package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoswagger "github.com/swaggo/echo-swagger"
	"net/http"
	"proman-backend/api/handler/auth"
	"proman-backend/api/handler/code"
	"proman-backend/api/handler/me"
	"proman-backend/api/handler/option"
	"proman-backend/api/handler/project"
	"proman-backend/api/handler/schedule"
	"proman-backend/api/handler/task"
	"proman-backend/api/handler/user"
	"proman-backend/config"
	"proman-backend/docs"
	"proman-backend/internal/database"
	"proman-backend/internal/pkg/file"
	"proman-backend/internal/pkg/log"
	"proman-backend/version"
	"strings"
)

// @title Proman Backend
// @description Proman Backend API
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @securityDefinitions.basic BasicAuth
// @in header

const appName = "Proman Backend"

func main() {
	defer log.RecoverWithTrace()

	e := echo.New()
	log.SetLogger(e)

	//git_api.InitGitlab()
	db := database.ConnectMongo()
	var err error

	//projects, _, err := git_api.Client.Users.CreateUser(
	//	&gitlab.CreateUserOptions{
	//		Email:    gitlab.Ptr("fawifanifaw"),
	//	})
	//if err != nil {
	//	log.Fatal("Gitlab projects error: ", err)
	//}
	//for _, project := range projects {
	//	fmt.Printf(project.Name)
	//}

	file.Sess, err = session.NewSession(&aws.Config{
		Endpoint: &config.S3.EndPoint,
		Region:   &config.S3.Region,
	})
	if err != nil {
		log.Fatal("Amazon S3 session error: ", err)
	}
	file.Uploader = s3manager.NewUploader(file.Sess)
	file.Downloader = s3manager.NewDownloader(file.Sess)
	file.S3Client = s3.New(file.Sess)

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     strings.Split(config.App.AllowOrigins, ","),
		AllowCredentials: true,
	}))

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="en">
    <head>
        <title>Proman Backend</title>
    </head>
    <body>
  		<h1>Welcome to %v</h1>
  		<p><a href="/api/version">version: %v</a></p>
  		<p><a href="/swagger/index.html#/">docs</a></p>
	</body>
	</html>`, appName, version.Version))
	})
	e.GET("/api/version", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"name":    appName,
			"version": version.Version,
		})
	})
	e.GET("/swagger/*", echoswagger.WrapHandler)

	docs.SwaggerInfo.Version = version.Version
	docs.SwaggerInfo.Host = config.App.SwaggerHost

	auth.NewHandler(e, db)
	me.NewHandler(e, db)
	project.NewHandler(e, db)
	task.NewHandler(e, db)
	user.NewHandler(e, db)
	schedule.NewHandler(e, db)
	code.NewHandler(e, db)
	option.NewHandler(e, db)

	log.Fatal(e.Start(fmt.Sprintf(`%v:%v`, config.App.Host, config.App.Port)))
}
