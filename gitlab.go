package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/xanzy/go-gitlab"
)

// GitLabClient wraps the GitLab API client
type GitLabClient struct {
	client *gitlab.Client
	ctx    context.Context
}

// NewGitLabClient creates a new GitLab client
func NewGitLabClient() (*GitLabClient, error) {
	ctx := context.Background()
	
	// Get token from environment
	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITLAB_TOKEN environment variable not set")
	}

	// Create GitLab client
	client, err := gitlab.NewClient(token)
	if err != nil {
		return nil, err
	}

	return &GitLabClient{
		client: client,
		ctx:    ctx,
	}, nil
}

// GetPipelineRuns retrieves pipeline runs for a project
func (g *GitLabClient) GetPipelineRuns(projectID string, limit int) ([]WorkflowRun, error) {
	pipelines, _, err := g.client.Pipelines.ListProjectPipelines(
		projectID,
		&gitlab.ListProjectPipelinesOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: limit,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	var workflowRuns []WorkflowRun
	for _, pipeline := range pipelines {
		workflowRun := WorkflowRun{
			ID:         fmt.Sprintf("%d", pipeline.ID),
			Project:    projectID,
			Workflow:   pipeline.Ref,
			Status:     string(pipeline.Status),
			Conclusion: string(pipeline.Status), // GitLab uses status for both
			CreatedAt:  *pipeline.CreatedAt,
			UpdatedAt:  *pipeline.UpdatedAt,
			URL:        pipeline.WebURL,
			Platform:   "gitlab",
			Branch:     pipeline.Ref,
			Commit:     pipeline.SHA,
			TriggeredBy: "system", // GitLab doesn't always have user info
		}
		workflowRuns = append(workflowRuns, workflowRun)
	}

	return workflowRuns, nil
}

// GetPipelineJobs retrieves jobs for a specific pipeline
func (g *GitLabClient) GetPipelineJobs(projectID string, pipelineID string) ([]Job, error) {
	pipelineIDInt, err := strconv.Atoi(pipelineID)
	if err != nil {
		return nil, err
	}
	
	jobs, _, err := g.client.Jobs.ListPipelineJobs(
		projectID,
		pipelineIDInt,
		&gitlab.ListJobsOptions{},
	)
	if err != nil {
		return nil, err
	}

	var jobList []Job
	for _, job := range jobs {
		jobItem := Job{
			ID:         fmt.Sprintf("%d", job.ID),
			RunID:      pipelineID,
			Name:       job.Name,
			Status:     string(job.Status),
			Conclusion: string(job.Status),
			URL:        job.WebURL,
		}

		// Add timing information
		if job.StartedAt != nil {
			startedAt := *job.StartedAt
			jobItem.StartedAt = &startedAt
		}
		if job.FinishedAt != nil {
			completedAt := *job.FinishedAt
			jobItem.CompletedAt = &completedAt
		}

		// GitLab doesn't have steps in the same way as GitHub Actions
		// We'll create a single step representing the job
		step := Step{
			Name:       job.Name,
			Status:     string(job.Status),
			Conclusion: string(job.Status),
		}
		if job.StartedAt != nil {
			startedAt := *job.StartedAt
			step.StartedAt = &startedAt
		}
		if job.FinishedAt != nil {
			completedAt := *job.FinishedAt
			step.CompletedAt = &completedAt
		}
		jobItem.Steps = append(jobItem.Steps, step)

		jobList = append(jobList, jobItem)
	}

	return jobList, nil
}

// GetPipelines retrieves available pipeline configurations
func (g *GitLabClient) GetPipelines(projectID string) ([]string, error) {
	// GitLab doesn't have a direct equivalent to GitHub's workflow list
	// We'll return the available branches that have pipelines
	branches, _, err := g.client.Branches.ListBranches(
		projectID,
		&gitlab.ListBranchesOptions{},
	)
	if err != nil {
		return nil, err
	}

	var pipelineNames []string
	for _, branch := range branches {
		pipelineNames = append(pipelineNames, branch.Name)
	}

	return pipelineNames, nil
}

// TriggerPipeline triggers a pipeline for a specific ref
func (g *GitLabClient) TriggerPipeline(projectID, ref string, variables map[string]string) error {
	// Convert variables to GitLab format
	var gitlabVars []*gitlab.PipelineVariableOptions
	for key, value := range variables {
		gitlabVars = append(gitlabVars, &gitlab.PipelineVariableOptions{
			Key:   &key,
			Value: &value,
		})
	}
	
	_, _, err := g.client.Pipelines.CreatePipeline(
		projectID,
		&gitlab.CreatePipelineOptions{
			Ref:       &ref,
			Variables: &gitlabVars,
		},
	)
	return err
}
