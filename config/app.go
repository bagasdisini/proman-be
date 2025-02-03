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

var Basic struct {
	Username string `mapstructure:"BASIC_AUTH_USERNAME"`
	Password string `mapstructure:"BASIC_AUTH_PASSWORD"`
}

var JWT struct {
	Key    string `mapstructure:"AUTH_JWT_KEY"`
	Expire int    `mapstructure:"AUTH_JWT_EXPIRE"`
}

var Mail struct {
	Enable     bool   `mapstructure:"MAIL_ENABLE"`
	Host       string `mapstructure:"MAIL_HOST"`
	Port       int    `mapstructure:"MAIL_PORT"`
	SenderName string `mapstructure:"MAIL_SENDER_NAME"`
	AuthMail   string `mapstructure:"MAIL_AUTH_EMAIL"`
	AuthPass   string `mapstructure:"MAIL_AUTH_PASSWORD"`
}

var Vcode struct {
	CheckEnable bool `mapstructure:"VCODE_CHECK_ENABLE"`
	Length      int  `mapstructure:"VCODE_LENGTH"`
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

	Basic.Username = os.Getenv("BASIC_AUTH_USERNAME")
	Basic.Password = os.Getenv("BASIC_AUTH_PASSWORD")

	if Basic.Username == "" {
		panic("BASIC_AUTH_USERNAME is not set")
	}
	if Basic.Password == "" {
		panic("BASIC_AUTH_PASSWORD is not set")
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

	mailEnable, err := strconv.ParseBool(os.Getenv("MAIL_ENABLE"))
	if err != nil {
		panic("MAIL_ENABLE is not valid")
	}
	Mail.Enable = mailEnable
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

	vcodeCheckEnable, err := strconv.ParseBool(os.Getenv("VCODE_CHECK_ENABLE"))
	if err != nil {
		panic("VCODE_CHECK_ENABLE is not valid")
	}
	Vcode.CheckEnable = vcodeCheckEnable
	vcodeLength, err := strconv.Atoi(os.Getenv("VCODE_LENGTH"))
	if err != nil {
		panic("VCODE_LENGTH is not valid")
	}
	if vcodeLength < 1 {
		panic("VCODE_LENGTH is not valid")
	}
	Vcode.Length = vcodeLength
}
