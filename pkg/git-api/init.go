package git_api

import (
	"github.com/xanzy/go-gitlab"
	"proman-backend/internal/config"
	"proman-backend/pkg/log"
)

var Client *gitlab.Client

func InitGitlab() {
	var err error
	Client, err = gitlab.NewClient(config.Gitlab.Token, gitlab.WithBaseURL(config.Gitlab.URL))
	if err != nil {
		log.Errorf("Failed to create client: %v", err)
		return
	}
}
