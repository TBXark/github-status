package render

import (
	"embed"
	_ "embed"
	"fmt"
	"github.com/TBXark/github-status/stats"
	"maps"
	"os"
	"slices"
	"strings"
	"text/template"
	"time"
)

//go:embed templates/overview.gohtml
var overviewSVG string

//go:embed templates/languages.gohtml
var languagesSVG string

//go:embed icons/*.svg
var iconsFS embed.FS

type SVGData string

func (s SVGData) WriteToPath(path string) error {
	return os.WriteFile(path, []byte(s), 0644)
}

type OverviewItem struct {
	Icon  string
	Name  string
	Value string
}

func loadIcon(name string) string {
	f, err := iconsFS.ReadFile("icons/" + name + ".svg")
	if err != nil {
		return ""
	}
	return string(f)
}

func OverviewSVG(animation bool, data *stats.Stats) (SVGData, error) {
	var input struct {
		Name      string
		Animation bool
		Items     []OverviewItem
	}
	input.Name = data.Name
	input.Animation = animation
	input.Items = append(input.Items, OverviewItem{
		Icon:  loadIcon("star"),
		Name:  "Stars",
		Value: fmt.Sprintf("%d", data.Stargazers),
	})
	input.Items = append(input.Items, OverviewItem{
		Icon:  loadIcon("repo-forked"),
		Name:  "Forks",
		Value: fmt.Sprintf("%d", data.Forks),
	})
	if data.LineChange != nil {
		input.Items = append(input.Items, OverviewItem{
			Icon:  loadIcon("diff"),
			Name:  "Lines of code changed",
			Value: fmt.Sprintf("%d", data.LineChange.Additions+data.LineChange.Deletions),
		})
	} else {
		input.Items = append(input.Items, OverviewItem{
			Icon:  loadIcon("git-commit"),
			Name:  fmt.Sprintf("Total commits (%d)", time.Now().Year()),
			Value: fmt.Sprintf("%d", data.Contributions.TotalCommitContributions),
		})
	}
	if data.Views != nil {
		input.Items = append(input.Items, OverviewItem{
			Icon:  loadIcon("eye"),
			Name:  "Repository views (past two weeks)",
			Value: fmt.Sprintf("%d", data.Views.Count),
		})
	} else {
		input.Items = append(input.Items, OverviewItem{
			Icon:  loadIcon("git-pull-request"),
			Name:  fmt.Sprintf("Total pull requests (%d)", time.Now().Year()),
			Value: fmt.Sprintf("%d", data.Contributions.TotalPullRequestContributions),
		})
	}
	input.Items = append(input.Items, OverviewItem{
		Icon:  loadIcon("repo-push"),
		Name:  "All-time contributions",
		Value: fmt.Sprintf("%d", data.Contributions.TotalContributions),
	})
	input.Items = append(input.Items, OverviewItem{
		Icon:  loadIcon("repo"),
		Name:  "Repositories with contributions",
		Value: fmt.Sprintf("%d", len(data.Repos)),
	})

	tmpl, err := template.New("overview").Parse(overviewSVG)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	err = tmpl.Execute(&buf, input)
	if err != nil {
		return "", err
	}
	return SVGData(buf.String()), nil
}

func LanguagesSVG(animation bool, data *stats.Stats) (SVGData, error) {
	var input struct {
		Animation bool
		Languages []*stats.LanguageStats
	}
	funcMap := template.FuncMap{
		"AnimationDelay": func(i int) int {
			return i * 150
		},
		"Percent": func(i float64) string {
			return fmt.Sprintf("%.3f%%", i)
		},
	}
	tmpl, err := template.New("languages").Funcs(funcMap).Parse(languagesSVG)
	if err != nil {
		return "", err
	}
	input.Animation = animation
	input.Languages = slices.SortedFunc(maps.Values(data.Languages), func(s1 *stats.LanguageStats, s2 *stats.LanguageStats) int {
		return s2.Size - s1.Size
	})
	var buf strings.Builder
	err = tmpl.Execute(&buf, input)
	if err != nil {
		return "", err
	}
	return SVGData(buf.String()), nil
}
