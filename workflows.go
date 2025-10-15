package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	qc "github.com/bevelwork/quick_color"
)

// watchWorkflows displays running workflows across all projects
func watchWorkflows(ctx context.Context, config *Config) {
	if len(config.Projects) == 0 {
		fmt.Printf("%s No projects tracked. Use 'quick_workflow add .' to add a project.\n", qc.Colorize("Info:", qc.ColorCyan))
		return
	}

	fmt.Printf("%s\n", qc.Colorize("Watching workflows across all projects...", qc.ColorBlue))
	fmt.Println()

	// Collect all workflow runs
	var allRuns []WorkflowRun
	for _, project := range config.Projects {
		runs, err := getWorkflowRunsForProject(ctx, project, 10)
		if err != nil {
			fmt.Printf("%s Failed to get workflows for %s: %v\n", qc.Colorize("Error:", qc.ColorRed), project.Name, err)
			continue
		}
		allRuns = append(allRuns, runs...)
	}

	if len(allRuns) == 0 {
		fmt.Printf("%s No workflow runs found\n", qc.Colorize("Info:", qc.ColorCyan))
		return
	}

	// Sort by creation time (newest first)
	sort.Slice(allRuns, func(i, j int) bool {
		return allRuns[i].CreatedAt.After(allRuns[j].CreatedAt)
	})

	// Display workflow runs
	displayWorkflowRuns(allRuns)

	// Allow user to select a run for details
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s", qc.Colorize("Select a workflow run for details (number or 'q' to quit): ", qc.ColorYellow))
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	input = strings.TrimSpace(input)
	
	if input == "q" || input == "" {
		return
	}

	runIndex, err := strconv.Atoi(input)
	if err != nil || runIndex < 1 || runIndex > len(allRuns) {
		fmt.Println("Invalid selection")
		return
	}

	selectedRun := allRuns[runIndex-1]
	showWorkflowDetails(ctx, config, selectedRun)
}

// startWorkflow allows starting a new workflow
func startWorkflow(ctx context.Context, config *Config, args []string) {
	if len(config.Projects) == 0 {
		fmt.Printf("%s No projects tracked. Use 'quick_workflow add .' to add a project.\n", qc.Colorize("Info:", qc.ColorCyan))
		return
	}

	// Select project
	selectedProject := selectProject(config)
	if selectedProject == nil {
		return
	}

	// Get available workflows
	workflows, err := getAvailableWorkflows(ctx, *selectedProject)
	if err != nil {
		fmt.Printf("%s Failed to get workflows: %v\n", qc.Colorize("Error:", qc.ColorRed), err)
		return
	}

	if len(workflows) == 0 {
		fmt.Printf("%s No workflows available for %s\n", qc.Colorize("Info:", qc.ColorCyan), selectedProject.Name)
		return
	}

	// Select workflow
	selectedWorkflow := selectWorkflow(workflows)
	if selectedWorkflow == "" {
		return
	}

	// Trigger workflow
	err = triggerWorkflow(ctx, *selectedProject, selectedWorkflow)
	if err != nil {
		fmt.Printf("%s Failed to trigger workflow: %v\n", qc.Colorize("Error:", qc.ColorRed), err)
		return
	}

	fmt.Printf("%s Triggered workflow '%s' for %s\n", qc.Colorize("Success:", qc.ColorGreen), selectedWorkflow, selectedProject.Name)
}

// listWorkflows shows historical workflow runs
func listWorkflows(ctx context.Context, config *Config, args []string) {
	if len(config.Projects) == 0 {
		fmt.Printf("%s No projects tracked. Use 'quick_workflow add .' to add a project.\n", qc.Colorize("Info:", qc.ColorCyan))
		return
	}

	// Parse limit from args
	limit := 20
	if len(args) > 0 {
		if l, err := strconv.Atoi(args[0]); err == nil {
			limit = l
		}
	}

	fmt.Printf("%s\n", qc.Colorize("Recent workflow runs:", qc.ColorBlue))
	fmt.Println()

	// Collect all workflow runs
	var allRuns []WorkflowRun
	for _, project := range config.Projects {
		runs, err := getWorkflowRunsForProject(ctx, project, limit)
		if err != nil {
			fmt.Printf("%s Failed to get workflows for %s: %v\n", qc.Colorize("Error:", qc.ColorRed), project.Name, err)
			continue
		}
		allRuns = append(allRuns, runs...)
	}

	if len(allRuns) == 0 {
		fmt.Printf("%s No workflow runs found\n", qc.Colorize("Info:", qc.ColorCyan))
		return
	}

	// Sort by creation time (newest first)
	sort.Slice(allRuns, func(i, j int) bool {
		return allRuns[i].CreatedAt.After(allRuns[j].CreatedAt)
	})

	// Display workflow runs
	displayWorkflowRuns(allRuns)
}

// getWorkflowRunsForProject retrieves workflow runs for a specific project
func getWorkflowRunsForProject(ctx context.Context, project Project, limit int) ([]WorkflowRun, error) {
	switch project.Platform {
	case "github":
		client, err := NewGitHubClient()
		if err != nil {
			return nil, err
		}
		return client.GetWorkflowRuns(project.Owner, project.Repo, limit)
	case "gitlab":
		client, err := NewGitLabClient()
		if err != nil {
			return nil, err
		}
		// For GitLab, we need to find the project ID first
		// This is a simplified approach - in practice, you'd want to store the project ID
		return client.GetPipelineRuns(project.Name, limit)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", project.Platform)
	}
}

// getAvailableWorkflows retrieves available workflows for a project
func getAvailableWorkflows(ctx context.Context, project Project) ([]string, error) {
	switch project.Platform {
	case "github":
		client, err := NewGitHubClient()
		if err != nil {
			return nil, err
		}
		return client.GetWorkflows(project.Owner, project.Repo)
	case "gitlab":
		client, err := NewGitLabClient()
		if err != nil {
			return nil, err
		}
		return client.GetPipelines(project.Name)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", project.Platform)
	}
}

// triggerWorkflow triggers a workflow for a project
func triggerWorkflow(ctx context.Context, project Project, workflowName string) error {
	switch project.Platform {
	case "github":
		client, err := NewGitHubClient()
		if err != nil {
			return err
		}
		// For GitHub, we need to get the workflow file name
		// This is simplified - in practice, you'd want to map workflow names to file names
		return client.TriggerWorkflow(project.Owner, project.Repo, workflowName, "main", nil)
	case "gitlab":
		client, err := NewGitLabClient()
		if err != nil {
			return err
		}
		return client.TriggerPipeline(project.Name, workflowName, nil)
	default:
		return fmt.Errorf("unsupported platform: %s", project.Platform)
	}
}

// displayWorkflowRuns displays a list of workflow runs
func displayWorkflowRuns(runs []WorkflowRun) {
	longestProject := 0
	for _, run := range runs {
		if len(run.Project) > longestProject {
			longestProject = len(run.Project)
		}
	}

	for i, run := range runs {
		// Alternate row colors
		rowColor := qc.AlternatingColor(i, qc.ColorWhite, qc.ColorCyan)
		
		// Color code the status
		statusColor := colorWorkflowStatus(run.Status, run.Conclusion)
		
		// Format time
		timeStr := run.CreatedAt.Format("2006-01-02 15:04")
		
		entry := fmt.Sprintf(
			"%3d. %-*s %-20s %s [%s] %s",
			i+1, longestProject, run.Project, run.Workflow,
			timeStr, qc.Colorize(run.Status, statusColor),
			run.Branch,
		)
		fmt.Println(qc.Colorize(entry, rowColor))
	}
}

// showWorkflowDetails displays detailed information about a workflow run
func showWorkflowDetails(ctx context.Context, config *Config, run WorkflowRun) {
	fmt.Printf("\n%s\n", qc.Colorize("Workflow Details:", qc.ColorBlue))
	fmt.Printf("Project: %s\n", qc.ColorizeBold(run.Project, qc.ColorGreen))
	fmt.Printf("Workflow: %s\n", run.Workflow)
	fmt.Printf("Status: %s\n", qc.Colorize(run.Status, colorWorkflowStatus(run.Status, run.Conclusion)))
	fmt.Printf("Branch: %s\n", run.Branch)
	fmt.Printf("Commit: %s\n", run.Commit)
	fmt.Printf("Created: %s\n", run.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("URL: %s\n", run.URL)
	fmt.Println()

	// Get jobs for this run
	jobs, err := getJobsForRun(ctx, run)
	if err != nil {
		fmt.Printf("%s Failed to get jobs: %v\n", qc.Colorize("Error:", qc.ColorRed), err)
		return
	}

	if len(jobs) == 0 {
		fmt.Printf("%s No jobs found for this run\n", qc.Colorize("Info:", qc.ColorCyan))
		return
	}

	// Display jobs
	fmt.Printf("%s\n", qc.Colorize("Jobs:", qc.ColorBlue))
	for i, job := range jobs {
		rowColor := qc.AlternatingColor(i, qc.ColorWhite, qc.ColorCyan)
		statusColor := colorJobStatus(job.Status, job.Conclusion)
		
		entry := fmt.Sprintf(
			"  %3d. %-30s [%s]",
			i+1, job.Name,
			qc.Colorize(job.Status, statusColor),
		)
		fmt.Println(qc.Colorize(entry, rowColor))
	}
}

// getJobsForRun retrieves jobs for a specific workflow run
func getJobsForRun(ctx context.Context, run WorkflowRun) ([]Job, error) {
	// Parse the project name to extract owner/repo and platform
	parts := strings.Split(run.Project, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid project format: %s (expected owner/repo)", run.Project)
	}
	
	owner := parts[0]
	repo := parts[1]
	
	// Create a temporary project for API calls
	project := Project{
		Name:     run.Project,
		Owner:    owner,
		Repo:     repo,
		Platform: run.Platform,
	}

	switch project.Platform {
	case "github":
		client, err := NewGitHubClient()
		if err != nil {
			return nil, err
		}
		return client.GetWorkflowJobs(project.Owner, project.Repo, run.ID)
	case "gitlab":
		client, err := NewGitLabClient()
		if err != nil {
			return nil, err
		}
		return client.GetPipelineJobs(project.Name, run.ID)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", project.Platform)
	}
}

// selectProject allows user to select a project
func selectProject(config *Config) *Project {
	if len(config.Projects) == 1 {
		return &config.Projects[0]
	}

	fmt.Printf("%s\n", qc.Colorize("Select a project:", qc.ColorBlue))
	for i, project := range config.Projects {
		rowColor := qc.AlternatingColor(i, qc.ColorWhite, qc.ColorCyan)
		platformColor := colorPlatform(project.Platform)
		
		entry := fmt.Sprintf(
			"%3d. %-30s [%s]",
			i+1, project.Name,
			qc.Colorize(project.Platform, platformColor),
		)
		fmt.Println(qc.Colorize(entry, rowColor))
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s", qc.Colorize("Select project (number): ", qc.ColorYellow))
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil
	}
	input = strings.TrimSpace(input)
	
	index, err := strconv.Atoi(input)
	if err != nil || index < 1 || index > len(config.Projects) {
		return nil
	}

	return &config.Projects[index-1]
}

// selectWorkflow allows user to select a workflow
func selectWorkflow(workflows []string) string {
	if len(workflows) == 1 {
		return workflows[0]
	}

	fmt.Printf("%s\n", qc.Colorize("Select a workflow:", qc.ColorBlue))
	for i, workflow := range workflows {
		rowColor := qc.AlternatingColor(i, qc.ColorWhite, qc.ColorCyan)
		entry := fmt.Sprintf("%3d. %s", i+1, workflow)
		fmt.Println(qc.Colorize(entry, rowColor))
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s", qc.Colorize("Select workflow (number): ", qc.ColorYellow))
	input, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	input = strings.TrimSpace(input)
	
	index, err := strconv.Atoi(input)
	if err != nil || index < 1 || index > len(workflows) {
		return ""
	}

	return workflows[index-1]
}

// colorWorkflowStatus returns a color for workflow status
func colorWorkflowStatus(status, conclusion string) string {
	switch status {
	case "completed":
		if conclusion == "success" {
			return qc.ColorGreen
		} else if conclusion == "failure" {
			return qc.ColorRed
		} else if conclusion == "cancelled" {
			return qc.ColorYellow
		}
		return qc.ColorWhite
	case "in_progress", "running":
		return qc.ColorBlue
	case "queued", "pending":
		return qc.ColorYellow
	case "failed":
		return qc.ColorRed
	default:
		return qc.ColorWhite
	}
}

// colorJobStatus returns a color for job status
func colorJobStatus(status, conclusion string) string {
	return colorWorkflowStatus(status, conclusion)
}
