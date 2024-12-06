package config

import (
	"os"
	"strings"
)

type Config struct {
	Login       string
	UserName    string
	AccessToken string

	ExcludeRepos []string
	ExcludeLangs []string
	IncludeOwner []string

	IgnorePrivateRepos       bool
	IgnoreForkedRepos        bool
	IgnoreArchivedRepos      bool
	IgnoreContributedToRepos bool

	WebhookURL string
}

func NewConfig(tokenValidate func(token string) bool) *Config {
	accessToken := os.Getenv("ACCESS_TOKEN")
	if !tokenValidate(accessToken) {
		accessToken = os.Getenv("GITHUB_TOKEN")
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

		WebhookURL: os.Getenv("WEBHOOK_URL"),
	}

	if len(conf.IncludeOwner) == 0 {
		conf.IncludeOwner = []string{conf.UserName}
	}

	return conf
}
