package query

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	ErrAcceptButNotReady = fmt.Errorf("request accepted but not ready")
	ErrTooManyRequests   = fmt.Errorf("too many requests")
)

type Queries struct {
	accessToken string
	client      *http.Client
}

func NewQueries(accessToken string) *Queries {
	return &Queries{
		accessToken: accessToken,
		client:      &http.Client{},
	}
}

func (q *Queries) IsValid() bool {
	if q.accessToken == "" {
		return false
	}
	_, err := q.requestRest(context.Background(), "user", nil)
	if err != nil {
		return false
	}
	return true
}

func (q *Queries) requestGraphql(ctx context.Context, query string) (json.RawMessage, error) {

	reqBody, err := json.Marshal(graphQLRequest{Query: query})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.github.com/graphql", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+q.accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := q.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var graphqlResp graphQLResponse
	err = json.Unmarshal(body, &graphqlResp)
	if err != nil {
		return nil, err
	}
	if len(graphqlResp.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %s", graphqlResp.Errors[0].Message)
	}
	return graphqlResp.Data, nil
}

func (q *Queries) requestRest(ctx context.Context, path string, params map[string]string) (json.RawMessage, error) {

	baseURL := fmt.Sprintf("https://api.github.com/%s", strings.TrimLeft(path, "/"))
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", q.accessToken))

	if params != nil {
		query := req.URL.Query()
		for key, value := range params {
			query.Add(key, value)
		}
		req.URL.RawQuery = query.Encode()
	}

	resp, err := q.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusAccepted {
		return nil, ErrAcceptButNotReady
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrTooManyRequests
	}
	return io.ReadAll(resp.Body)
}

func (q *Queries) formatAfterCursor(after string) string {
	if after == "" {
		return "null"
	}
	return fmt.Sprintf(`"%s"`, after)
}

func (q *Queries) repositoriesQuery(query, login string) string {
	return fmt.Sprintf(`
query {
  user(login: "%s") {
    repositories: %s {
      pageInfo {
        hasNextPage
        endCursor
      }
      nodes {
        nameWithOwner
        stargazers {
          totalCount
        }
        forkCount
        isFork
        isArchived
        isPrivate
        languages(first: 10, orderBy: {field: SIZE, direction: DESC}) {
          edges {
            size
            node {
              name
              color
            }
          }
        }
      }
    }
  }
}`, login, strings.TrimSpace(query))
}

func (q *Queries) Repositories(ctx context.Context, login, after string) (*RepositoriesPage, error) {
	query := q.repositoriesQuery(fmt.Sprintf(`
	repositories(
        first: 100,
        orderBy: {
            field: UPDATED_AT,
            direction: DESC
        },
        isFork: false,
        after: %s
    )`, q.formatAfterCursor(after)), login)
	data, err := sendQuery[Repositories](ctx, q, query)
	if err != nil {
		return nil, err
	}
	return &data.Repositories, nil
}

func (q *Queries) RepositoriesContributedTo(ctx context.Context, login, after string) (*RepositoriesPage, error) {
	query := q.repositoriesQuery(fmt.Sprintf(`
	repositoriesContributedTo(
        first: 100,
        includeUserRepositories: false,
        orderBy: {
            field: UPDATED_AT,
            direction: DESC
        },
        contributionTypes: [
            COMMIT,
            PULL_REQUEST,
            REPOSITORY,
            PULL_REQUEST_REVIEW
        ]
        after: %s
    )`, q.formatAfterCursor(after)), login)
	data, err := sendQuery[Repositories](ctx, q, query)
	if err != nil {
		return nil, err
	}
	return &data.Repositories, nil
}

func (q *Queries) ContributionsCollection(ctx context.Context, login string) (*ContributionsCollection, error) {
	query := fmt.Sprintf(`
query {
  user(login: "%s") {
    contributionsCollection {
      contributionYears
      totalCommitContributions
      totalIssueContributions
      totalPullRequestContributions
      totalPullRequestReviewContributions
    }
  }
}`, login)
	return sendQuery[ContributionsCollection](ctx, q, query)
}

func (q *Queries) AllContribYears(ctx context.Context, login string, years []int) (AllContribYears, error) {
	var byYears string
	for _, year := range years {
		byYears += fmt.Sprintf(`
    	year%d: contributionsCollection(from: "%d-01-01T00:00:00Z",to: "%d-01-01T00:00:00Z") {
    	  contributionCalendar {
    	    totalContributions
    	  }
    	}`, year, year, year+1)
	}
	query := fmt.Sprintf(`
query {
  user(login: "%s") {
    %s
  }
}`, login, byYears)
	data, err := sendQuery[AllContribYears](ctx, q, query)
	if err != nil {
		return nil, err
	}
	return *data, nil
}

func (q *Queries) RepoTraffic(ctx context.Context, repo string) (*RepoTraffic, error) {
	return sendRequest[RepoTraffic](ctx, q, fmt.Sprintf("/repos/%s/traffic/views", repo), 1, nil)
}

func (q *Queries) RepoContributors(ctx context.Context, repo string) (*[]RepoContributor, error) {
	return sendRequest[[]RepoContributor](ctx, q, fmt.Sprintf("/repos/%s/stats/contributors", repo), 60, nil)
}

func sendQuery[T any](ctx context.Context, client *Queries, query string) (*T, error) {
	data, err := client.requestGraphql(ctx, query)
	if err != nil {
		return nil, err
	}
	var result struct {
		User T `json:"user"`
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return &result.User, nil
}

func sendRequest[T any](ctx context.Context, client *Queries, path string, maxTries int, params map[string]string) (*T, error) {
	for i := 0; i < maxTries; i++ {
		data, err := client.requestRest(ctx, path, params)
		if errors.Is(err, ErrAcceptButNotReady) {
			log.Printf("Request accepted but not ready, retrying in 1 second")
			time.Sleep(time.Second * 1)
			continue
		}
		if errors.Is(err, ErrTooManyRequests) {
			log.Printf("Too many requests, retrying in 2 second")
			time.Sleep(time.Second * 2)
			continue
		}
		if err != nil {
			return nil, err
		}
		var result T
		err = json.Unmarshal(data, &result)
		if err != nil {
			return nil, err
		}
		return &result, nil
	}
	return nil, fmt.Errorf("max tries exceeded")
}

type (
	graphQLRequest struct {
		Query string `json:"query"`
	}
	graphQLResponse struct {
		Data   json.RawMessage `json:"data,omitempty"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors,omitempty"`
	}
)

type (
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
		Stargazers    struct {
			TotalCount int `json:"totalCount"`
		} `json:"stargazers"`
		ForkCount  int  `json:"forkCount"`
		IsFork     bool `json:"isFork"`
		IsArchived bool `json:"isArchived"`
		IsPrivate  bool `json:"isPrivate"`
		Languages  struct {
			Edges []struct {
				Size int `json:"size"`
				Node struct {
					Name  string `json:"name"`
					Color string `json:"color"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"languages"`
	}
	RepositoriesPage struct {
		PageInfo struct {
			HasNextPage bool   `json:"hasNextPage"`
			EndCursor   string `json:"endCursor"`
		} `json:"pageInfo"`
		Nodes []Repository `json:"nodes"`
	}
	Repositories struct {
		Repositories RepositoriesPage `json:"repositories"`
	}
)

type (
	ContributionsCollection struct {
		ContributionsCollection struct {
			ContributionYears                   []int `json:"contributionYears"`
			TotalCommitContributions            int   `json:"totalCommitContributions"`
			TotalIssueContributions             int   `json:"totalIssueContributions"`
			TotalPullRequestContributions       int   `json:"totalPullRequestContributions"`
			TotalPullRequestReviewContributions int   `json:"totalPullRequestReviewContributions"`
		} `json:"contributionsCollection"`
	}
	ContributionCalendar struct {
		ContributionCalendar struct {
			TotalContributions int `json:"TotalContributions"`
		} `json:"contributionCalendar"`
	}
	AllContribYears = map[string]ContributionCalendar
)

type (
	RepoContributor struct {
		Total int `json:"total"`
		Weeks []struct {
			W int `json:"w"`
			A int `json:"a"`
			D int `json:"d"`
			C int `json:"c"`
		} `json:"weeks"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
	}

	RepoTraffic struct {
		Count   int `json:"count"`
		Uniques int `json:"uniques"`
		Views   []struct {
			Timestamp time.Time `json:"timestamp"`
			Count     int       `json:"count"`
			Uniques   int       `json:"uniques"`
		} `json:"views"`
	}
)
