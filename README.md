# Quick Workflow

A simple Go CLI tool for monitoring GitHub Actions and GitLab CI workflows across multiple projects. This tool allows you to add projects, watch running workflows, start new workflows, and review historical runs from a unified interface.

Part of the `Quick Tools` family of tools from [Bevel Work](https://bevel.work/quick-tools).

## ✨ Features

- **Multi-Platform Support**: Monitor both GitHub Actions and GitLab CI workflows
- **Project Management**: Add and track multiple repositories
- **Live Monitoring**: Watch running workflows across all projects
- **Workflow Triggering**: Start new workflows from the command line
- **Historical Review**: List and review past workflow runs
- **Unified Interface**: Standardized view across different CI platforms
- **Interactive Selection**: Easy navigation with numbered menus
- **Color-Coded Output**: Visual status indicators and alternating row colors

## Installation

### Required Software

1. **Go 1.24.4 or later** - [Download and install Go](https://golang.org/dl/)
2. **Git** - For repository detection and remote URL parsing

### Install with Go
```bash
go install github.com/bevelwork/quick_workflow@latest
quick_workflow --version
```

### Or Build from Source
```bash
git clone https://github.com/bevelwork/quick_workflow.git
cd quick_workflow
go build -o quick_workflow .
./quick_workflow --version
```

## Configuration

### Environment Variables

Set up API tokens for the platforms you want to use:

```bash
# For GitHub Actions
export GITHUB_TOKEN=your_github_token_here

# For GitLab CI
export GITLAB_TOKEN=your_gitlab_token_here
```

### State File

The tool stores tracked projects in `~/.config/quick_workflow/state.json`. This file is created automatically when you add your first project.

## Usage

### Basic Commands

```bash
# Add current directory as a project
quick_workflow add .

# Add a specific repository
quick_workflow add /path/to/repository

# Watch running workflows across all projects
quick_workflow watch

# Start a new workflow
quick_workflow start

# List recent workflow runs
quick_workflow list

# List tracked projects
quick_workflow projects

# Remove a project
quick_workflow remove project_name

# Show help
quick_workflow help
```

### Examples

```bash
# Add current repository and watch workflows
quick_workflow add .
quick_workflow watch

# List last 50 workflow runs
quick_workflow list 50

# Start a deployment workflow
quick_workflow start
```

## How It Works

1. **Project Detection**: Automatically detects GitHub and GitLab repositories from git remote URLs
2. **API Integration**: Uses GitHub and GitLab APIs to fetch workflow information
3. **Unified Data Model**: Standardizes workflow runs, jobs, and steps across platforms
4. **State Management**: Tracks projects and their configurations in a JSON state file
5. **Interactive Interface**: Provides numbered menus for easy selection and navigation

## Supported Platforms

### GitHub Actions
- List workflow runs
- View job details and steps
- Trigger workflow dispatches
- Monitor status and conclusions

### GitLab CI
- List pipeline runs
- View job details
- Trigger new pipelines
- Monitor status and conclusions

## API Token Requirements

### GitHub
Your `GITHUB_TOKEN` needs the following permissions:
- `repo` (for private repositories)
- `actions:read` (to read workflow runs)
- `actions:write` (to trigger workflows)

### GitLab
Your `GITLAB_TOKEN` needs the following scopes:
- `read_api` (to read pipeline information)
- `api` (to trigger pipelines)

## Troubleshooting

### Common Issues

1. **"GITHUB_TOKEN environment variable not set"**
   - Set your GitHub token: `export GITHUB_TOKEN=your_token`
   - Ensure the token has the required permissions

2. **"GITLAB_TOKEN environment variable not set"**
   - Set your GitLab token: `export GITLAB_TOKEN=your_token`
   - Ensure the token has the required scopes

3. **"Current directory is not a git repository"**
   - Run the command from within a git repository
   - Or use `quick_workflow add /path/to/repo` to specify a path

4. **"Failed to parse remote URL"**
   - Ensure your git remote URL is in a supported format
   - Supported formats: `https://github.com/owner/repo.git`, `git@github.com:owner/repo.git`

5. **"No projects tracked"**
   - Add a project first: `quick_workflow add .`
   - Check that the project was added: `quick_workflow projects`

### Git Remote URL Formats

The tool supports these remote URL formats:

**GitHub:**
- `https://github.com/owner/repo.git`
- `git@github.com:owner/repo.git`

**GitLab:**
- `https://gitlab.com/owner/repo.git`
- `git@gitlab.com:owner/repo.git`
- Custom GitLab instances (e.g., `https://git.example.com/owner/repo.git`)

## Development

### Building from Source

```bash
git clone https://github.com/bevelwork/quick_workflow.git
cd quick_workflow
go mod download
go build -o quick_workflow .
```

### Running Tests

```bash
go test -v ./...
```

### Code Formatting

```bash
go fmt ./...
go vet ./...
```

## Version Management

This project uses a date-based versioning system: `major.minor.YYYYMMDD`

- **Major**: Breaking changes or major feature additions
- **Minor**: New features or significant improvements
- **Date**: Build date in YYYYMMDD format

## License

Apache 2.0

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Related Tools

- [Quick SSM](https://github.com/bevelwork/quick_ssm) - Connect to AWS EC2 instances via SSM
- [Quick ECS](https://github.com/bevelwork/quick_ecs) - Manage AWS ECS services
- [Quick Color](https://github.com/bevelwork/quick_color) - Color utilities for CLI tools
