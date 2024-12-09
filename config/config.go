package config

import (
	"os"
	"strings"
)

type Config struct {
	UserName    string `json:"user_name"`
	AccessToken string `json:"access_token"`

	ExcludeRepos []string `json:"exclude_repos"`
	ExcludeLangs []string `json:"exclude_langs"`
	IncludeOwner []string `json:"include_owner"`

	IgnorePrivateRepos       bool `json:"ignore_private_repos"`
	IgnoreForkedRepos        bool `json:"ignore_forked_repos"`
	IgnoreArchivedRepos      bool `json:"ignore_archived_repos"`
	IgnoreContributedToRepos bool `json:"ignore_contributed_to_repos"`

	IgnoreLinesChanged bool `json:"ignore_lines_changed"`
	IgnoreRepoViews    bool `json:"ignore_repo_views"`

	WebhookURL string `json:"webhook_url"`
}

func NewConfig(tokenValidate func(token string) bool) *Config {
	accessToken := os.Getenv("ACCESS_TOKEN")
	if !tokenValidate(accessToken) {
		accessToken = os.Getenv("GITHUB_TOKEN")
		if !tokenValidate(accessToken) {
			return nil
		}
	}

	userName := os.Getenv("CUSTOM_ACTOR")
	if userName == "" {
		userName = os.Getenv("GITHUB_ACTOR")
	}

	stringSliceFromEnv := func(key string) []string {
		if value := os.Getenv(key); value != "" {
			return strings.Split(value, ",")
		}
		return []string{}
	}

	boolFromEnv := func(key string) bool {
		return os.Getenv(key) == "true"
	}

	conf := &Config{
		UserName:    userName,
		AccessToken: accessToken,

		ExcludeRepos: stringSliceFromEnv("EXCLUDE_REPOS"),
		ExcludeLangs: stringSliceFromEnv("EXCLUDE_LANGS"),
		IncludeOwner: stringSliceFromEnv("INCLUDE_OWNER"),

		IgnorePrivateRepos:       boolFromEnv("IGNORE_PRIVATE_REPOS"),
		IgnoreForkedRepos:        boolFromEnv("IGNORE_FORKED_REPOS"),
		IgnoreArchivedRepos:      boolFromEnv("IGNORE_ARCHIVED_REPOS"),
		IgnoreContributedToRepos: boolFromEnv("IGNORE_CONTRIBUTED_TO_REPOS"),

		IgnoreLinesChanged: boolFromEnv("IGNORE_LINES_CHANGED"),
		IgnoreRepoViews:    boolFromEnv("IGNORE_REPO_VIEWS"),

		WebhookURL: os.Getenv("WEBHOOK_URL"),
	}

	if len(conf.IncludeOwner) == 0 {
		conf.IncludeOwner = []string{conf.UserName}
	}

	return conf
}
