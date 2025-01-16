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
gh-rerun-actions-workflow -repo "<owner>/<repository>" -run-id <run-id> -max-reruns <limit> -interval <duration>
```

### Example

```bash
gh-rerun-actions-workflow -repo "octocat/hello-world" -run-id 12345678 -max-reruns 3 -interval 30s
```

## Options

- `-repo`: Target repository (`owner/repository`).
- `-run-id`: Workflow run ID.
- `-max-reruns`: Max retries (default: 5).
- `-interval`: Check interval (default: 1 minute).

## License

MIT License.
