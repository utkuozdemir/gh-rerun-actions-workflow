# gh-rerun-actions-workflow

A simple Go tool to monitor and rerun failed GitHub Actions workflows.

## Features

- Monitors workflow runs in a GitHub repository.
- Reruns failed jobs up to a configurable limit.
- Supports configurable check intervals.

## Install

```bash
go install github.com/utkuozdemir/gh-rerun-actions-workflow@latest
```

## Usage

```bash
export GITHUB_TOKEN="<your-github-token>"
gh-rerun-actions-workflow -repo "<owner>/<repository>" -run-id <run-id> -max-reruns <limit> -interval <duration>
```

### Example

```bash
export GITHUB_TOKEN=ghp_mySuperSecretToken
gh-rerun-actions-workflow -repo "utkuozdemir/gh-rerun-actions-workflow" -run-id 12345678 -max-reruns 3 -interval 30s
```

## Options

- `-repo`: Target repository (`owner/repository`).
- `-run-id`: Workflow run ID.
- `-max-reruns`: Max retries (default: 5).
- `-interval`: Check interval (default: 1 minute).
- `-api-call-timeout`: Timeout for GitHub API calls (default: 5 seconds).

## License

MIT License.
