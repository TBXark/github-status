package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"github.com/TBXark/github-status/config"
	"github.com/TBXark/github-status/query"
	"github.com/TBXark/github-status/render"
	"github.com/TBXark/github-status/stats"
	"log"
	"net/http"
	"os"
)

func main() {
	output := flag.String("output", "output", "The output directory")
	debug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	conf := config.NewConfig(func(token string) bool {
		return query.NewQueries(token).IsValid()
	})
	if conf == nil {
		log.Fatalf("Invalid config")
	}
	loader := stats.NewStats(
		conf.UserName,
		conf.AccessToken,
		stats.IgnoreForkedRepos(conf.IgnoreForkedRepos),
		stats.IgnoreArchivedRepos(conf.IgnoreArchivedRepos),
		stats.IgnorePrivateRepos(conf.IgnorePrivateRepos),
		stats.IgnoreContributedToRepos(conf.IgnoreContributedToRepos),
		stats.ExcludeRepos(conf.ExcludeRepos...),
		stats.ExcludeLangs(conf.ExcludeLangs...),
		stats.IncludeOwner(conf.IncludeOwner...),
	)
	stat, err := loader.GetStats(context.Background())
	if err != nil {
		log.Fatalf("Failed to get stats: %v", err)
	}
	if e := saveStat(stat, *output); e != nil {
		log.Printf("Failed to save stat: %v", e)
	}
	if e := sendWebhook(conf, stat); e != nil {
		log.Printf("Failed to send webhook: %v", e)
	}
	if *debug {
		data, _ := json.MarshalIndent(stat, "", "  ")
		_ = os.WriteFile(*output+"/data.json", data, 0644)
	}
}

func saveStat(stat *stats.Stats, output string) error {
	err := os.MkdirAll(output, 0755)
	if err != nil {
		return err
	}

	overview, err := render.OverviewSVG(stat)
	if err != nil {
		return err
	}
	err = overview.WriteToPath(output + "/overview.svg")
	if err != nil {
		return err
	}

	languages, err := render.LanguagesSVG(stat)
	if err != nil {
		return err
	}
	err = languages.WriteToPath(output + "/languages.svg")
	if err != nil {
		return err
	}
	return nil
}

func sendWebhook(conf *config.Config, obj any) error {
	if conf.WebhookURL == "" {
		return nil
	}
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", conf.WebhookURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
