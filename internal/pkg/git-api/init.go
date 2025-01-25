package git_api

import (
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"proman-backend/config"
	"proman-backend/internal/pkg/log"
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
