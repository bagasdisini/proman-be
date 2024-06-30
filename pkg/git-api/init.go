package git_api

import (
	"github.com/xanzy/go-gitlab"
	"proman-backend/internal/config"
	"proman-backend/pkg/log"
)

var Client *gitlab.Client

func InitGitlab() {
	var err error
	Client, err = gitlab.NewClient(config.Gitlab.Token)
	if err != nil {
		log.Errorf("Failed to create client: %v", err)
		return
	}
}
