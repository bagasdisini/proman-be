package config

import "os"

var Gitlab struct {
	Token string `mapstructure:"GITLAB_ACCESS_TOKEN"`
}

func initGitlab() {
	Gitlab.Token = os.Getenv("GITLAB_ACCESS_TOKEN")

	if Gitlab.Token == "" {
		panic("GITLAB_ACCESS_TOKEN is required")
	}
}
