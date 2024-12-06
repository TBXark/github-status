# GitHub Status Generator

A tool to generate beautiful SVG cards showing your GitHub statistics and programming language distribution.

## Features

- Generates an overview SVG card with GitHub statistics
- Creates a languages SVG card showing your programming language distribution
- Customizable through environment variables
- Supports excluding specific repositories and languages
- Option to ignore forked and archived repositories
- Flexible configuration for multiple GitHub owners

## Installation

```bash
go install github.com/tbxark/github-status@latest
```

## Usage

### Basic Usage

1. Set up your GitHub token as an environment variable:
```bash
export GITHUB_TOKEN=your_github_token
export CUSTOM_ACTOR=your_github_username
```

2. Run the program:
```bash
github-status --output ./output
```

### Environment Variables

The following environment variables can be used to customize the behavior:

- `GITHUB_TOKEN` or `ACCESS_TOKEN`: Your GitHub personal access token
- `GITHUB_ACTOR` or `CUSTOM_ACTOR`: GitHub username
- `EXCLUDE_REPOS`: Comma-separated list of repositories to exclude
- `EXCLUDE_LANGS`: Comma-separated list of languages to exclude
- `INCLUDE_OWNER`: Comma-separated list of GitHub usernames to include (defaults to your username)
- `IGNORE_FORKED_REPOS`: Set to "true" to ignore forked repositories
- `IGNORE_ARCHIVED_REPOS`: Set to "true" to ignore archived repositories

### Command Line Options

- `--output`: Specify the output directory for generated SVG files (default: "output")

## Output Files

The program generates three files in the output directory:

1. `overview.svg`: Contains general GitHub statistics
2. `languages.svg`: Shows programming language distribution
3. `stats.json`: Raw statistics data in JSON format

## Example

```bash
export GITHUB_TOKEN=ghp_your_token_here
export EXCLUDE_REPOS=repo1,repo2
export EXCLUDE_LANGS=HTML,CSS
export IGNORE_FORKED_REPOS=true
github-status --output ./my-stats
```

## Thanks

This project was inspired by the [jstrieb/github-stats](https://github.com/jstrieb/github-stats) project


## License

**github-status** is released under the MIT license. [See LICENSE](LICENSE) for details.