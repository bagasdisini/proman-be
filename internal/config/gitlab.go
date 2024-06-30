package config

import "os"

var Gitlab struct {
	Token string `mapstructure:"GITLAB_ACCESS_TOKEN"`
	URL   string `mapstructure:"GITLAB_URL"`
}

func initGitlab() {
	Gitlab.Token = os.Getenv("GITLAB_ACCESS_TOKEN")

	if Gitlab.Token == "" {
		panic("GITLAB_ACCESS_TOKEN is required")
	}
	if Gitlab.URL == "" {
		panic("GITLAB_URL is required")
	}
}
