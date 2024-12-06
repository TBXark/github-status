package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/tbxark/github-status/config"
	"github.com/tbxark/github-status/render"
	"github.com/tbxark/github-status/stats"
	"log"
	"os"
)

func main() {
	output := flag.String("output", "output", "The output directory")
	flag.Parse()

	conf := config.NewConf()
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
	err = saveStat(stat, *output)
	if err != nil {
		log.Fatalf("Failed to save stats: %v", err)
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

	file, err := os.Create("output/stats.json")
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(stat)
	if err != nil {
		return err
	}

	return nil
}
