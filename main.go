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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/ssh/terminal"
	"net/http"
	"os"
	"proman-backend/api/handler"
	"proman-backend/api/repository"
	"proman-backend/docs"
	"proman-backend/internal/config"
	"proman-backend/internal/database"
	_const "proman-backend/pkg/const"
	"proman-backend/pkg/file"
	git_api "proman-backend/pkg/git-api"
	"proman-backend/pkg/log"
	"proman-backend/pkg/util"
	"proman-backend/version"
	"strings"
	"syscall"
	"time"
)

// @title Proman Backend
// @description Proman Backend API
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
const appName = "Proman Backend"

func main() {
	defer log.RecoverWithTrace()

	e := echo.New()
	log.SetLogger(e)

	git_api.InitGitlab()
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

	if len(os.Args) > 1 {
		if os.Args[1] == "init-admin" {
			initAdmin(db)
			return
		}
	}

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

	handler.NewAuthHandler(e, db)
	handler.NewMeHandler(e, db)
	handler.NewProjectHandler(e, db)
	handler.NewTaskHandler(e, db)
	handler.NewUserHandler(e, db)
	handler.NewScheduleHandler(e, db)

	log.Fatal(e.Start(fmt.Sprintf(`%v:%v`, config.App.Host, config.App.Port)))
}

func initAdmin(db *mongo.Database) {
	repos := repository.NewUserRepository(db)

	if repos.Check() {
		fmt.Println("Initiate failed, user already exists.")
		return
	}

	fmt.Println("This command will creates a users with role admin")
	fmt.Print("Insert password for admin : ")
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	password := string(bytePassword)
	password = strings.TrimSuffix(password, "\n")

	user := &repository.User{
		ID:        primitive.NewObjectID(),
		Email:     "admin@admin.com",
		Password:  util.CryptPassword(password),
		Name:      "admin",
		Role:      _const.RoleAdmin,
		Position:  _const.PositionSysAdmin,
		CreatedAt: time.Now(),
		IsDeleted: false,
	}

	_, err := repos.Insert(user)
	if err != nil {
		fmt.Println("\nInitiate super admin failed, there is an error : ", err)
		return
	}

	fmt.Println("\n\nInitiate admin success!")
	fmt.Println("Admin email : admin@admin.com")
}
