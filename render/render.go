package render

import (
	_ "embed"
	"github.com/TBXark/github-status/stats"
	"maps"
	"os"
	"slices"
	"strings"
	"text/template"
)

//go:embed templates/overview.gohtml
var overviewSVG string

//go:embed templates/languages.gohtml
var languagesSVG string

type SVGData string

func (s SVGData) WriteToPath(path string) error {
	return os.WriteFile(path, []byte(s), 0644)
}

func OverviewSVG(data *stats.Stats) (SVGData, error) {
	tmpl, err := template.New("overview").Parse(overviewSVG)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return SVGData(buf.String()), nil
}

func LanguagesSVG(data *stats.Stats) (SVGData, error) {
	funcMap := template.FuncMap{
		"AnimationDelay": func(i int) int {
			return i * 150
		},
	}
	tmpl, err := template.New("languages").Funcs(funcMap).Parse(languagesSVG)
	if err != nil {
		return "", err
	}
	langList := slices.SortedFunc(maps.Values(data.Languages), func(s1 *stats.LanguageStats, s2 *stats.LanguageStats) int {
		return s2.Size - s1.Size
	})
	var buf strings.Builder
	err = tmpl.Execute(&buf, langList)
	if err != nil {
		return "", err
	}
	return SVGData(buf.String()), nil
}
