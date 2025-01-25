package config

import (
	"os"
	"strconv"
)

var App struct {
	Host         string `mapstructure:"APP_HOST"`
	Port         string `mapstructure:"APP_PORT"`
	SwaggerHost  string `mapstructure:"SWAGGER_HOST"`
	AllowOrigins string `mapstructure:"CORS_ALLOW_ORIGINS"`
}

var JWT struct {
	Key    string `mapstructure:"AUTH_JWT_KEY"`
	Expire int    `mapstructure:"AUTH_JWT_EXPIRE"`
}

var Mail struct {
	Host       string `mapstructure:"MAIL_HOST"`
	Port       int    `mapstructure:"MAIL_PORT"`
	SenderName string `mapstructure:"MAIL_SENDER_NAME"`
	AuthMail   string `mapstructure:"MAIL_AUTH_EMAIL"`
	AuthPass   string `mapstructure:"MAIL_AUTH_PASSWORD"`
}

func initApp() {
	App.Host = os.Getenv("APP_HOST")
	App.Port = os.Getenv("APP_PORT")
	App.SwaggerHost = os.Getenv("SWAGGER_HOST")
	App.AllowOrigins = os.Getenv("CORS_ALLOW_ORIGINS")

	if App.Host == "" {
		panic("APP_HOST is not set")
	}
	if App.Port == "" {
		panic("APP_PORT is not set")
	}
	if App.SwaggerHost == "" {
		panic("SWAGGER_HOST is not set")
	}
	if App.AllowOrigins == "" {
		panic("CORS_ALLOW_ORIGINS is not set")
	}

	JWT.Key = os.Getenv("AUTH_JWT_KEY")
	expire, err := strconv.Atoi(os.Getenv("AUTH_JWT_EXPIRE"))
	if err != nil {
		panic("AUTH_JWT_EXPIRE is not valid")
	}
	JWT.Expire = expire

	if JWT.Key == "" {
		panic("AUTH_JWT_KEY is not set")
	}
	if JWT.Expire == 0 {
		JWT.Expire = 7
	} else if JWT.Expire < 0 {
		panic("AUTH_JWT_EXPIRE must be greater than 0")
	}

	Mail.Host = os.Getenv("MAIL_HOST")
	portInt, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
	if err != nil {
		panic("MAIL_PORT is not valid")
	}
	Mail.Port = portInt
	Mail.SenderName = os.Getenv("MAIL_SENDER_NAME")
	Mail.AuthMail = os.Getenv("MAIL_AUTH_EMAIL")
	Mail.AuthPass = os.Getenv("MAIL_AUTH_PASSWORD")

	if Mail.Host == "" {
		panic("MAIL_HOST is not set")
	}
	if Mail.Port < 1 {
		panic("MAIL_PORT is not set")
	}
	if Mail.SenderName == "" {
		panic("MAIL_SENDER_NAME is not set")
	}
	if Mail.AuthMail == "" {
		panic("MAIL_AUTH_EMAIL is not set")
	}
	if Mail.AuthPass == "" {
		panic("MAIL_AUTH_PASSWORD is not set")
	}
}
