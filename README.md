# GitHub Status Generator

A tool to generate beautiful SVG cards showing your GitHub statistics and programming language distribution.

[![overview](https://raw.githubusercontent.com/TBXark/TBXark/refs/heads/stats/overview.svg)](https://github.com/TBXark/TBXark) [![languages](https://raw.githubusercontent.com/TBXark/TBXark/refs/heads/stats/languages.svg)](https://github.com/TBXark/TBXark)


## Features

- Generates an overview SVG card with GitHub statistics:
  - Total contributions
  - Total stars and forks
  - Lines of code changed
  - Repository views
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

### Basic Usage

1. Set up your GitHub token as an environment variable:
```bash
export GITHUB_TOKEN=your_github_token  # Required: GitHub personal access token
export CUSTOM_ACTOR=your_github_username  # Optional: Defaults to GITHUB_ACTOR
```

2. Run the program:
```bash
github-status --output ./output
```

3. The generated SVG files will be available in the output directory.

### Environment Variables

The following environment variables can be used to customize the behavior:

- `GITHUB_TOKEN` or `ACCESS_TOKEN`: Your GitHub personal access token (required)
- `GITHUB_ACTOR` or `CUSTOM_ACTOR`: GitHub username
- `EXCLUDE_REPOS`: Comma-separated list of repositories to exclude (e.g., "repo1,repo2")
- `EXCLUDE_LANGS`: Comma-separated list of languages to exclude (e.g., "HTML,CSS")
- `INCLUDE_OWNER`: Comma-separated list of GitHub usernames to include (defaults to your username)
- `IGNORE_FORKED_REPOS`: Set to "true" to ignore forked repositories
- `IGNORE_ARCHIVED_REPOS`: Set to "true" to ignore archived repositories
- `IGNORE_PRIVATE_REPOS`: Set to "true" to ignore private repositories
- `IGNORE_CONTRIBUTED_TO_REPOS`: Set to "true" to ignore repositories you've contributed to
- `WEBHOOK_URL`: URL to send the generated statistics to (optional)

### Command Line Options

- `--output`: Specify the output directory for generated SVG files (default: "output")
- `--debug`: Enable debug mode to generate additional JSON output

## Output Files

The program generates three files in the output directory:

1. `overview.svg`: Contains general GitHub statistics
2. `languages.svg`: Shows programming language distribution
3. `stats.json`: Raw statistics data in JSON format

## Example

```bash
# Basic usage with token
export ACCESS_TOKEN=ghp_your_token_here
github-status --output ./my-stats

# Advanced usage with filters
export ACCESS_TOKEN=ghp_your_token_here
export EXCLUDE_REPOS=repo1,repo2
export EXCLUDE_LANGS=HTML,CSS
export IGNORE_FORKED_REPOS=true
export IGNORE_PRIVATE_REPOS=true
github-status --output ./my-stats --debug

# Using with multiple owners
export ACCESS_TOKEN=ghp_your_token_here
export INCLUDE_OWNER=owner1,owner2,owner3
github-status --output ./my-stats
```

## Integration

You can integrate the generated SVG files into your GitHub profile README by adding the following markdown:

```markdown
![GitHub Overview](./output/overview.svg)
![Top Languages](./output/languages.svg)
```

## Development

### Prerequisites

- Go 1.23 or higher
- GitHub Personal Access Token with the following scopes:
  - `repo`: Full access to private and public repositories
  - `read:user`: Read access to user profile data

### Building from Source

```bash
# Clone the repository
git clone https://github.com/tbxark/github-status.git
cd github-status

# Build the binary
make build-linux-amd  # For Linux AMD64
make build-linux-arm  # For Linux ARM64
```

## Thanks

This project was inspired by the [jstrieb/github-stats](https://github.com/jstrieb/github-stats) project

## License

**github-status** is released under the MIT license. [See LICENSE](LICENSE) for details.