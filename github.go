package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

// GitHubClient wraps the GitHub API client
type GitHubClient struct {
	client *github.Client
	ctx    context.Context
}

// NewGitHubClient creates a new GitHub client
func NewGitHubClient() (*GitHubClient, error) {
	ctx := context.Background()
	
	// Get token from environment
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	// Create OAuth2 client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	// Create GitHub client
	client := github.NewClient(tc)

	return &GitHubClient{
		client: client,
		ctx:    ctx,
	}, nil
}

// GetWorkflowRuns retrieves workflow runs for a repository
func (g *GitHubClient) GetWorkflowRuns(owner, repo string, limit int) ([]WorkflowRun, error) {
	runs, _, err := g.client.Actions.ListRepositoryWorkflowRuns(
		g.ctx,
		owner,
		repo,
		&github.ListWorkflowRunsOptions{
			ListOptions: github.ListOptions{
				PerPage: limit,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	var workflowRuns []WorkflowRun
	for _, run := range runs.WorkflowRuns {
		workflowRun := WorkflowRun{
			ID:         fmt.Sprintf("%d", run.GetID()),
			Project:    fmt.Sprintf("%s/%s", owner, repo),
			Workflow:   run.GetName(),
			Status:     run.GetStatus(),
			Conclusion: run.GetConclusion(),
			CreatedAt:  run.GetCreatedAt().Time,
			UpdatedAt:  run.GetUpdatedAt().Time,
			URL:        run.GetHTMLURL(),
			Platform:   "github",
			Branch:     run.GetHeadBranch(),
			Commit:     run.GetHeadSHA(),
			TriggeredBy: run.GetTriggeringActor().GetLogin(),
		}
		workflowRuns = append(workflowRuns, workflowRun)
	}

	return workflowRuns, nil
}

// GetWorkflowJobs retrieves jobs for a specific workflow run
func (g *GitHubClient) GetWorkflowJobs(owner, repo string, runID string) ([]Job, error) {
	runIDInt, err := strconv.ParseInt(runID, 10, 64)
	if err != nil {
		return nil, err
	}
	
	jobs, _, err := g.client.Actions.ListWorkflowJobs(
		g.ctx,
		owner,
		repo,
		runIDInt,
		&github.ListWorkflowJobsOptions{},
	)
	if err != nil {
		return nil, err
	}

	var jobList []Job
	for _, job := range jobs.Jobs {
		jobItem := Job{
			ID:         fmt.Sprintf("%d", job.GetID()),
			RunID:      fmt.Sprintf("%d", job.GetRunID()),
			Name:       job.GetName(),
			Status:     job.GetStatus(),
			Conclusion: job.GetConclusion(),
			URL:        job.GetHTMLURL(),
		}

		// Add timing information
		if job.StartedAt != nil {
			startedAt := job.StartedAt.Time
			jobItem.StartedAt = &startedAt
		}
		if job.CompletedAt != nil {
			completedAt := job.CompletedAt.Time
			jobItem.CompletedAt = &completedAt
		}

		// Add steps
		for _, step := range job.Steps {
			stepItem := Step{
				Name:       step.GetName(),
				Status:     step.GetStatus(),
				Conclusion: step.GetConclusion(),
			}
			if step.StartedAt != nil {
				startedAt := step.StartedAt.Time
				stepItem.StartedAt = &startedAt
			}
			if step.CompletedAt != nil {
				completedAt := step.CompletedAt.Time
				stepItem.CompletedAt = &completedAt
			}
			jobItem.Steps = append(jobItem.Steps, stepItem)
		}

		jobList = append(jobList, jobItem)
	}

	return jobList, nil
}

// GetWorkflows retrieves available workflows for a repository
func (g *GitHubClient) GetWorkflows(owner, repo string) ([]string, error) {
	workflows, _, err := g.client.Actions.ListWorkflows(
		g.ctx,
		owner,
		repo,
		&github.ListOptions{},
	)
	if err != nil {
		return nil, err
	}

	var workflowNames []string
	for _, workflow := range workflows.Workflows {
		workflowNames = append(workflowNames, workflow.GetName())
	}

	return workflowNames, nil
}

// TriggerWorkflow triggers a workflow dispatch
func (g *GitHubClient) TriggerWorkflow(owner, repo, workflowID, ref string, inputs map[string]string) error {
	// For now, we'll implement a simplified version that just returns an error
	// indicating that workflow triggering is not yet implemented
	return fmt.Errorf("workflow triggering not yet implemented for GitHub")
}
