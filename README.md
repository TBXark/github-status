# GitHub Status Generator

A tool to generate beautiful SVG cards showing your GitHub statistics and programming language distribution.

[![overview](https://raw.githubusercontent.com/TBXark/TBXark/refs/heads/stats/overview.svg)](https://github.com/TBXark/TBXark) [![languages](https://raw.githubusercontent.com/TBXark/TBXark/refs/heads/stats/languages.svg)](https://github.com/TBXark/TBXark)


## Features

- Generates an overview SVG card with GitHub statistics
- Creates a languages SVG card showing your programming language distribution
- Highly customizable through environment variables
- Supports excluding specific repositories and languages
- Smart filtering options for forked, archived, and private repositories
- Flexible configuration for multiple GitHub owners
- Webhook support for integration with other services

## Installation

```bash
go install github.com/TBXark/github-status@latest
```

## Usage

## Configuration

All configuration is done through environment variables. Here are all the available options:

| Environment Variable            | Type     | Description                                           | Default      |
|---------------------------------|----------|-------------------------------------------------------|--------------|
| `ACCESS_TOKEN` / `GITHUB_TOKEN` | string   | GitHub access token for API authentication            | Required     |
| `CUSTOM_ACTOR` / `GITHUB_ACTOR` | string   | GitHub username                                       | Required     |
| `EXCLUDE_REPOS`                 | string[] | Comma-separated list of repositories to exclude       | `[]`         |
| `EXCLUDE_LANGS`                 | string[] | Comma-separated list of languages to exclude          | `[]`         |
| `INCLUDE_OWNER`                 | string[] | Comma-separated list of GitHub owners to include      | `[username]` |
| `IGNORE_PRIVATE_REPOS`          | bool     | Whether to ignore private repositories                | `false`      |
| `IGNORE_FORKED_REPOS`           | bool     | Whether to ignore forked repositories                 | `false`      |
| `IGNORE_ARCHIVED_REPOS`         | bool     | Whether to ignore archived repositories               | `false`      |
| `IGNORE_CONTRIBUTED_TO_REPOS`   | bool     | Whether to ignore repositories you've contributed to  | `false`      |
| `IGNORE_LINES_CHANGED`          | bool     | Whether to ignore lines of code changed in statistics | `false`      |
| `IGNORE_REPO_VIEWS`             | bool     | Whether to ignore repository view counts              | `false`      |
| `WEBHOOK_URL`                   | string   | URL for webhook notifications                         | `""`         |

## Best Practices

Reference [My GitHub Action](https://github.com/TBXark/TBXark/blob/master/.github/workflows/update-status.yml)

Create a repository action to automatically update the stats. The action will update the stats and push the changes to the main branch. And you can use it svg card in your README.md.


## Thanks

This project was inspired by the [jstrieb/github-stats](https://github.com/jstrieb/github-stats) project

## License

**github-status** is released under the MIT license. [See LICENSE](LICENSE) for details.