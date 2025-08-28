package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/TBXark/github-status/config"
	"github.com/TBXark/github-status/query"
	"github.com/TBXark/github-status/render"
	"github.com/TBXark/github-status/stats"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	output := flag.String("output", "output", "The output directory")
	debug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	conf := config.NewConfig(func(token string) bool {
		return query.NewQueries(token).IsValid()
	})
	if conf == nil {
		return fmt.Errorf("invalid config")
	}
	loader := stats.NewStats(
		conf.UserName,
		conf.AccessToken,
		stats.IgnoreForkedRepos(conf.IgnoreForkedRepos),
		stats.IgnoreArchivedRepos(conf.IgnoreArchivedRepos),
		stats.IgnorePrivateRepos(conf.IgnorePrivateRepos),
		stats.IgnoreContributedToRepos(conf.IgnoreContributedToRepos),
		stats.IgnoreLinesChanged(conf.IgnoreLinesChanged),
		stats.IgnoreRepoViews(conf.IgnoreRepoViews),
		stats.ExcludeRepos(conf.ExcludeRepos...),
		stats.ExcludeLangs(conf.ExcludeLangs...),
		stats.IncludeOwner(conf.IncludeOwner...),
	)
	stat, err := loader.GetStats(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}
	if e := saveStat(conf.Animation, stat, *output); e != nil {
		log.Printf("Failed to save stat: %v", e)
	}
	if e := sendWebhook(conf, stat); e != nil {
		log.Printf("Failed to send webhook: %v", e)
	}
	if *debug {
		data, _ := json.MarshalIndent(stat, "", "  ")
		_ = os.WriteFile(*output+"/data.json", data, 0o644)
	}
	return nil
}

func saveStat(animation bool, stat *stats.Stats, output string) error {
	err := os.MkdirAll(output, 0o755)
	if err != nil {
		return err
	}

	overview, err := render.OverviewSVG(animation, stat)
	if err != nil {
		return err
	}
	err = overview.WriteToPath(output + "/overview.svg")
	if err != nil {
		return err
	}

	languages, err := render.LanguagesSVG(animation, stat)
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
	data, err := json.MarshalIndent(map[string]any{
		"status": obj,
		"config": conf,
	}, "", "  ")
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", conf.WebhookURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}
