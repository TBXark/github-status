package stats

import (
	"context"
	"fmt"
	"github.com/tbxark/github-status/query"
	"log"
	"strings"
	"sync"
)

type (
	Stats struct {
		Name       string `json:"name"`
		Stargazers int    `json:"stargazers"`
		Forks      int    `json:"forks"`

		Languages map[string]*LanguageStats `json:"languages"`
		Repos     map[string]*RepoStats     `json:"repos"`

		TotalContributions int `json:"totalContributions"`
		Additions          int `json:"additions"`
		Deletions          int `json:"deletions"`
		Views              int `json:"views"`
	}

	LanguageStats struct {
		Name        string  `json:"name"`
		Size        int     `json:"size"`
		Occurrences int     `json:"occurrences"`
		Color       string  `json:"color"`
		Proportion  float64 `json:"proportion"`
	}

	RepoStats struct {
		Name       string         `json:"name"`
		Forks      int            `json:"forks"`
		Stargazers int            `json:"stargazers"`
		Languages  map[string]int `json:"languages"`
		Ignored    bool           `json:"ignored"`
	}

	Filter struct {
		ignorePrivateRepos  bool
		ignoreForkedRepos   bool
		ignoreArchivedRepos bool

		excludeRepos map[string]struct{}
		excludeLangs map[string]struct{}
		includeOwner map[string]struct{}
	}
)

type Loader struct {
	username string
	filter   *Filter
	queries  *query.Queries
}

type Option func(*Loader)

func NewStats(username, accessToken string, options ...Option) *Loader {
	s := &Loader{
		username: username,
		filter: &Filter{
			excludeRepos: make(map[string]struct{}),
			excludeLangs: make(map[string]struct{}),
			includeOwner: make(map[string]struct{}),
		},
		queries: query.NewQueries(username, accessToken),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

func IgnoreForkedRepos(flag bool) Option {
	return func(s *Loader) {
		s.filter.ignoreForkedRepos = flag
	}
}

func IgnoreArchivedRepos(flag bool) Option {
	return func(s *Loader) {
		s.filter.ignoreArchivedRepos = flag
	}
}

func IgnorePrivateRepos(flag bool) Option {
	return func(s *Loader) {
		s.filter.ignorePrivateRepos = flag
	}
}

func ExcludeRepos(repos ...string) Option {
	return func(s *Loader) {
		for _, repo := range repos {
			s.filter.excludeRepos[strings.ToLower(repo)] = struct{}{}
		}
	}
}

func ExcludeLangs(langs ...string) Option {
	return func(s *Loader) {
		for _, lang := range langs {
			s.filter.excludeLangs[strings.ToLower(lang)] = struct{}{}
		}
	}
}

func IncludeOwner(owners ...string) Option {
	return func(s *Loader) {
		for _, owner := range owners {
			s.filter.includeOwner[strings.ToLower(owner)] = struct{}{}
		}
	}
}

func (s *Loader) GetStats(ctx context.Context) (*Stats, error) {

	stats := &Stats{
		Name:      s.username,
		Languages: make(map[string]*LanguageStats),
		Repos:     make(map[string]*RepoStats),
	}
	var nextOwned, nextContrib string

	var reqGroup sync.WaitGroup
	var readGroup sync.WaitGroup

	viewChan := make(chan int)
	linesChan := make(chan [2]int)
	semaphore := make(chan struct{}, 10)

	readGroup.Add(2)
	go func(r *Stats) {
		defer readGroup.Done()
		for views := range viewChan {
			r.Views += views
		}
	}(stats)
	go func(r *Stats) {
		defer readGroup.Done()
		for lines := range linesChan {
			r.Additions += lines[0]
			r.Deletions += lines[1]
		}
	}(stats)
	for {
		data, err := query.Query[query.ReposOverview](ctx, s.queries, query.BuildReposOverviewQuery(nextContrib, nextOwned))
		if err != nil {
			return nil, fmt.Errorf("query Repos overview: %w", err)
		}
		nodes := append(data.Repositories.Nodes, data.RepositoriesContributedTo.Nodes...)
		for _, repo := range nodes {
			if s.mergeRepoToStats(&repo, stats) == nil {
				continue
			}
			reqGroup.Add(1)
			go func(repo string) {
				semaphore <- struct{}{}
				defer func() {
					<-semaphore
					reqGroup.Done()
				}()
				if views, e := s.views(ctx, repo); e == nil {
					viewChan <- views
				}
				if lines, e := s.linesChanged(ctx, repo); e == nil {
					linesChan <- lines
				}
			}(repo.NameWithOwner)
		}
		if !data.Repositories.PageInfo.HasNextPage &&
			!data.RepositoriesContributedTo.PageInfo.HasNextPage {
			break
		}
		nextOwned = data.Repositories.PageInfo.EndCursor
		nextContrib = data.RepositoriesContributedTo.PageInfo.EndCursor
	}

	var totalSize int
	for _, lang := range stats.Languages {
		totalSize += lang.Size
	}
	for n, lang := range stats.Languages {
		lang.Proportion = 100 * float64(lang.Size) / float64(totalSize)
		stats.Languages[n] = lang
	}

	if totalContributions, e := s.totalContributions(ctx); e == nil {
		stats.TotalContributions = totalContributions
	}

	reqGroup.Wait()
	close(viewChan)
	close(linesChan)
	readGroup.Wait()

	return stats, nil
}

func (s *Loader) mergeRepoToStats(repo *query.Repository, stats *Stats) *RepoStats {
	if _, ok := stats.Repos[repo.NameWithOwner]; ok {
		log.Printf("Repo %s already exists in stats", repo.NameWithOwner)
		return nil
	}

	owner := strings.Split(repo.NameWithOwner, "/")[0]
	if _, ok := s.filter.includeOwner[strings.ToLower(owner)]; !ok {
		return nil
	}

	repoStat := &RepoStats{
		Name:       repo.NameWithOwner,
		Forks:      repo.ForkCount,
		Stargazers: repo.Stargazers.TotalCount,
		Languages:  make(map[string]int),
		Ignored:    true,
	}

	stats.Stargazers += repo.Stargazers.TotalCount
	stats.Forks += repo.ForkCount
	stats.Repos[repo.NameWithOwner] = repoStat

	if _, ok := s.filter.excludeRepos[strings.ToLower(repo.NameWithOwner)]; ok {
		return repoStat
	}

	if s.filter.ignoreForkedRepos && repo.IsFork {
		return repoStat
	}
	if s.filter.ignoreArchivedRepos && repo.IsArchived {
		return repoStat
	}
	if s.filter.ignorePrivateRepos && repo.IsPrivate {
		return repoStat
	}

	for _, lang := range repo.Languages.Edges {
		repoStat.Languages[lang.Node.Name] = lang.Size
		if _, ok := s.filter.excludeLangs[strings.ToLower(lang.Node.Name)]; ok {
			continue
		}
		if stats.Languages[lang.Node.Name] == nil {
			stats.Languages[lang.Node.Name] = &LanguageStats{
				Name:  lang.Node.Name,
				Color: lang.Node.Color,
			}
		}
		stats.Languages[lang.Node.Name].Size += lang.Size
		stats.Languages[lang.Node.Name].Occurrences += 1
	}
	repoStat.Ignored = false
	return repoStat
}

func (s *Loader) totalContributions(ctx context.Context) (int, error) {
	yearsData, err := query.Query[query.ContribYears](ctx, s.queries, query.BuildContribYearsQuery())
	if err != nil {
		return 0, err
	}
	allContrib, err := query.Query[query.AllContribYears](ctx, s.queries, query.BuildAllContribQuery(yearsData.ContributionsCollection.ContributionYears))
	if err != nil {
		return 0, err
	}

	total := 0
	for _, year := range *allContrib {
		total += year.ContributionCalendar.TotalContributions
	}
	return total, nil
}

func (s *Loader) linesChanged(ctx context.Context, repo string) ([2]int, error) {
	username := strings.ToLower(s.username)
	additions, deletions := 0, 0
	con, err := query.Request[[]query.RepoContributor](ctx, s.queries, fmt.Sprintf("/repos/%s/stats/contributors", repo), nil)
	if err != nil {
		return [2]int{0, 0}, err
	}
	for _, contributor := range *con {
		if strings.ToLower(contributor.Author.Login) != username {
			continue
		}
		for _, week := range contributor.Weeks {
			additions += week.A
			deletions += week.D
		}
	}
	return [2]int{additions, deletions}, nil
}

func (s *Loader) views(ctx context.Context, repo string) (int, error) {
	total := 0
	traffic, err := query.Request[query.RepoTraffic](ctx, s.queries, fmt.Sprintf("/repos/%s/traffic/views", repo), nil)
	if err != nil {
		return 0, err
	}
	for _, view := range traffic.Views {
		total += view.Count
	}
	return total, nil
}

func (r Stats) LinesChanged() int {
	return r.Additions + r.Deletions
}

func (r Stats) ReposCount() int {
	return len(r.Repos)
}

func (l LanguageStats) Percent() string {
	return fmt.Sprintf("%.3f%%", l.Proportion)
}
