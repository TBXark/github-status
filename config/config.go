package config

import (
	"net/http"
	"os"
	"strings"
)

type Conf struct {
	UserName    string
	AccessToken string

	ExcludeRepos []string
	ExcludeLangs []string
	IncludeOwner []string

	IgnorePrivateRepos  bool
	IgnoreForkedRepos   bool
	IgnoreArchivedRepos bool
}

func isGithubAccessTokenValid(accessToken string) bool {
	if accessToken == "" {
		return false
	}
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func NewConf() *Conf {
	accessToken := os.Getenv("ACCESS_TOKEN")
	if !isGithubAccessTokenValid(accessToken) {
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

	conf := &Conf{
		UserName:    userName,
		AccessToken: accessToken,

		ExcludeRepos: stringSliceFromEnv("EXCLUDE_REPOS"),
		ExcludeLangs: stringSliceFromEnv("EXCLUDE_LANGS"),
		IncludeOwner: stringSliceFromEnv("INCLUDE_OWNER"),

		IgnorePrivateRepos:  os.Getenv("IGNORE_PRIVATE_REPOS") == "true",
		IgnoreForkedRepos:   os.Getenv("IGNORE_FORKED_REPOS") == "true",
		IgnoreArchivedRepos: os.Getenv("IGNORE_ARCHIVED_REPOS") == "true",
	}

	if len(conf.IncludeOwner) == 0 {
		conf.IncludeOwner = []string{conf.UserName}
	}

	return conf
}
