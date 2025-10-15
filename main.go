// Package main provides a command-line tool for monitoring GitHub Actions and GitLab CI workflows.
// The tool allows users to add projects, watch running workflows, start new workflows, and list historical runs.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	qc "github.com/bevelwork/quick_color"
	versionpkg "github.com/bevelwork/quick_workflow/version"
)

// Project represents a tracked project with its repository information
type Project struct {
	Name        string `json:"name"`
	Owner       string `json:"owner"`
	Repo        string `json:"repo"`
	Platform    string `json:"platform"` // "github" or "gitlab"
	RemoteURL   string `json:"remote_url"`
	AddedAt     string `json:"added_at"`
	AccessToken string `json:"access_token,omitempty"` // Optional access token
}

// WorkflowRun represents a unified workflow run across platforms
type WorkflowRun struct {
	ID          string    `json:"id"`
	Project     string    `json:"project"`
	Workflow    string    `json:"workflow"`
	Status      string    `json:"status"`
	Conclusion  string    `json:"conclusion"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	URL         string    `json:"url"`
	Platform    string    `json:"platform"`
	Branch      string    `json:"branch"`
	Commit      string    `json:"commit"`
	TriggeredBy string    `json:"triggered_by"`
}

// Job represents a job within a workflow run
type Job struct {
	ID        string    `json:"id"`
	RunID     string    `json:"run_id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Conclusion string   `json:"conclusion"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Steps     []Step    `json:"steps"`
	URL       string    `json:"url"`
}

// Step represents a step within a job
type Step struct {
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	Conclusion  string     `json:"conclusion"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Logs        string     `json:"logs,omitempty"`
}

// Config holds application configuration
type Config struct {
	StateFile string
	Projects  []Project
}

// version is set at build time via ldflags
var version = ""

func main() {
	// Parse command line flags
	showVersion := flag.Bool("version", false, "Show version information")
	stateFile := flag.String("state", "", "Path to state file (default: ~/.config/quick_workflow/state.json)")
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Println(resolveVersion())
		os.Exit(0)
	}

	// Set default state file if not provided
	if *stateFile == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("Failed to get user home directory:", err)
		}
		*stateFile = filepath.Join(homeDir, ".config", "quick_workflow", "state.json")
	}

	// Ensure state directory exists
	stateDir := filepath.Dir(*stateFile)
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		log.Fatal("Failed to create state directory:", err)
	}

	config := &Config{
		StateFile: *stateFile,
	}

	// Load existing projects
	if err := loadProjects(config); err != nil {
		log.Printf("Warning: Failed to load projects: %v", err)
		config.Projects = []Project{}
	}

	// Parse command
	args := flag.Args()
	if len(args) == 0 {
		showHelp()
		return
	}

	command := args[0]
	remainingArgs := args[1:]

	ctx := context.Background()

	switch command {
	case "add":
		if len(remainingArgs) == 0 {
			// Add current directory
			addCurrentProject(ctx, config)
		} else {
			// Add specific project
			addProject(ctx, config, remainingArgs[0])
		}
	case "watch":
		watchWorkflows(ctx, config)
	case "start":
		startWorkflow(ctx, config, remainingArgs)
	case "list":
		listWorkflows(ctx, config, remainingArgs)
	case "projects":
		listProjects(config)
	case "remove":
		if len(remainingArgs) == 0 {
			fmt.Println("Usage: quick_workflow remove <project_name>")
			return
		}
		removeProject(config, remainingArgs[0])
	case "help":
		showHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		showHelp()
	}
}

// resolveVersion returns the version string
func resolveVersion() string {
	if strings.TrimSpace(version) != "" {
		return version
	}
	if strings.TrimSpace(versionpkg.Full) != "" {
		return versionpkg.Full
	}
	return fmt.Sprintf("v%d.%d.%s", versionpkg.Major, versionpkg.Minor, "unknown")
}

// showHelp displays help information
func showHelp() {
	fmt.Printf("%s\n", qc.Colorize("Quick Workflow - Monitor GitHub Actions and GitLab CI workflows", qc.ColorBlue))
	fmt.Println()
	fmt.Printf("%s\n", qc.Colorize("Usage:", qc.ColorYellow))
	fmt.Println("  quick_workflow <command> [options]")
	fmt.Println()
	fmt.Printf("%s\n", qc.Colorize("Commands:", qc.ColorYellow))
	fmt.Println("  add [path]     Add current directory or specified path as a project")
	fmt.Println("  watch          Watch running workflows across all projects")
	fmt.Println("  start          Start a new workflow")
	fmt.Println("  list           List historical workflow runs")
	fmt.Println("  projects       List tracked projects")
	fmt.Println("  remove <name>  Remove a project from tracking")
	fmt.Println("  help           Show this help message")
	fmt.Println()
	fmt.Printf("%s\n", qc.Colorize("Examples:", qc.ColorYellow))
	fmt.Println("  quick_workflow add .                    # Add current directory")
	fmt.Println("  quick_workflow add /path/to/repo         # Add specific repository")
	fmt.Println("  quick_workflow watch                     # Watch running workflows")
	fmt.Println("  quick_workflow start                     # Start a new workflow")
	fmt.Println("  quick_workflow list                      # List recent workflow runs")
	fmt.Println("  quick_workflow projects                  # List tracked projects")
	fmt.Println()
	fmt.Printf("%s\n", qc.Colorize("Configuration:", qc.ColorYellow))
	fmt.Println("  State file: ~/.config/quick_workflow/state.json")
	fmt.Println("  Set GITHUB_TOKEN for GitHub API access")
	fmt.Println("  Set GITLAB_TOKEN for GitLab API access")
}

// addCurrentProject adds the current directory as a project
func addCurrentProject(ctx context.Context, config *Config) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get current directory:", err)
	}

	// Check if we're in a git repository
	if !isGitRepository(cwd) {
		log.Fatal("Current directory is not a git repository")
	}

	// Get remote URL
	remoteURL, err := getGitRemoteURL(cwd)
	if err != nil {
		log.Fatal("Failed to get git remote URL:", err)
	}

	// Parse remote URL to determine platform and owner/repo
	platform, owner, repo, err := parseRemoteURL(remoteURL)
	if err != nil {
		log.Fatal("Failed to parse remote URL:", err)
	}

	// Create project
	project := Project{
		Name:      fmt.Sprintf("%s/%s", owner, repo),
		Owner:     owner,
		Repo:      repo,
		Platform:  platform,
		RemoteURL: remoteURL,
		AddedAt:   time.Now().Format(time.RFC3339),
	}

	// Check if project already exists
	for _, existing := range config.Projects {
		if existing.Name == project.Name {
			fmt.Printf("%s Project %s is already tracked\n", qc.Colorize("Info:", qc.ColorCyan), qc.ColorizeBold(project.Name, qc.ColorGreen))
			return
		}
	}

	// Add project
	config.Projects = append(config.Projects, project)
	if err := saveProjects(config); err != nil {
		log.Fatal("Failed to save project:", err)
	}

	fmt.Printf("%s Added project: %s (%s)\n", qc.Colorize("Success:", qc.ColorGreen), qc.ColorizeBold(project.Name, qc.ColorGreen), platform)
}

// addProject adds a specific project
func addProject(ctx context.Context, config *Config, path string) {
	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal("Failed to resolve path:", err)
	}

	// Check if it's a git repository
	if !isGitRepository(absPath) {
		log.Fatal("Path is not a git repository:", absPath)
	}

	// Get remote URL
	remoteURL, err := getGitRemoteURL(absPath)
	if err != nil {
		log.Fatal("Failed to get git remote URL:", err)
	}

	// Parse remote URL
	platform, owner, repo, err := parseRemoteURL(remoteURL)
	if err != nil {
		log.Fatal("Failed to parse remote URL:", err)
	}

	// Create project
	project := Project{
		Name:      fmt.Sprintf("%s/%s", owner, repo),
		Owner:     owner,
		Repo:      repo,
		Platform:  platform,
		RemoteURL: remoteURL,
		AddedAt:   time.Now().Format(time.RFC3339),
	}

	// Check if project already exists
	for _, existing := range config.Projects {
		if existing.Name == project.Name {
			fmt.Printf("%s Project %s is already tracked\n", qc.Colorize("Info:", qc.ColorCyan), qc.ColorizeBold(project.Name, qc.ColorGreen))
			return
		}
	}

	// Add project
	config.Projects = append(config.Projects, project)
	if err := saveProjects(config); err != nil {
		log.Fatal("Failed to save project:", err)
	}

	fmt.Printf("%s Added project: %s (%s)\n", qc.Colorize("Success:", qc.ColorGreen), qc.ColorizeBold(project.Name, qc.ColorGreen), platform)
}


// listProjects shows tracked projects
func listProjects(config *Config) {
	if len(config.Projects) == 0 {
		fmt.Printf("%s No projects tracked. Use 'quick_workflow add .' to add a project.\n", qc.Colorize("Info:", qc.ColorCyan))
		return
	}

	fmt.Printf("%s\n", qc.Colorize("Tracked Projects:", qc.ColorBlue))
	fmt.Println()

	for i, project := range config.Projects {
		// Alternate row colors
		rowColor := qc.AlternatingColor(i, qc.ColorWhite, qc.ColorCyan)
		
		// Color code the platform
		platformColor := colorPlatform(project.Platform)
		
		entry := fmt.Sprintf(
			"%3d. %-30s %s [%s]",
			i+1, project.Name, project.RemoteURL,
			qc.Colorize(project.Platform, platformColor),
		)
		fmt.Println(qc.Colorize(entry, rowColor))
	}
}

// removeProject removes a project from tracking
func removeProject(config *Config, name string) {
	for i, project := range config.Projects {
		if project.Name == name {
			// Remove project
			config.Projects = append(config.Projects[:i], config.Projects[i+1:]...)
			if err := saveProjects(config); err != nil {
				log.Fatal("Failed to save projects:", err)
			}
			fmt.Printf("%s Removed project: %s\n", qc.Colorize("Success:", qc.ColorGreen), qc.ColorizeBold(name, qc.ColorGreen))
			return
		}
	}
	fmt.Printf("%s Project not found: %s\n", qc.Colorize("Error:", qc.ColorRed), name)
}

// Helper functions

// isGitRepository checks if a directory is a git repository
func isGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// getGitRemoteURL gets the remote URL from git
func getGitRemoteURL(path string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// parseRemoteURL parses a git remote URL to extract platform, owner, and repo
func parseRemoteURL(url string) (platform, owner, repo string, err error) {
	// Handle different URL formats
	if strings.HasPrefix(url, "https://github.com/") {
		platform = "github"
		parts := strings.Split(strings.TrimPrefix(url, "https://github.com/"), "/")
		if len(parts) >= 2 {
			owner = parts[0]
			repo = strings.TrimSuffix(parts[1], ".git")
			return
		}
	} else if strings.HasPrefix(url, "git@github.com:") {
		platform = "github"
		parts := strings.Split(strings.TrimPrefix(url, "git@github.com:"), "/")
		if len(parts) >= 2 {
			owner = parts[0]
			repo = strings.TrimSuffix(parts[1], ".git")
			return
		}
	} else if strings.Contains(url, "gitlab") {
		platform = "gitlab"
		// Parse GitLab URL
		// This is a simplified parser - in practice, you'd want more robust parsing
		parts := strings.Split(url, "/")
		if len(parts) >= 2 {
			// Find the owner/repo part
			for i, part := range parts {
				if strings.Contains(part, "gitlab") && i+2 < len(parts) {
					owner = parts[i+1]
					repo = strings.TrimSuffix(parts[i+2], ".git")
					return
				}
			}
		}
	}
	
	return "", "", "", fmt.Errorf("unsupported remote URL format: %s", url)
}

// colorPlatform returns a color for the platform
func colorPlatform(platform string) string {
	switch platform {
	case "github":
		return qc.ColorPurple
	case "gitlab":
		return qc.ColorPurple
	default:
		return qc.ColorWhite
	}
}

