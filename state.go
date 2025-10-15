package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// State represents the application state
type State struct {
	Projects []Project `json:"projects"`
	Version  string    `json:"version"`
}

// loadProjects loads projects from the state file
func loadProjects(config *Config) error {
	// Check if state file exists
	if _, err := os.Stat(config.StateFile); os.IsNotExist(err) {
		// Create empty state
		config.Projects = []Project{}
		return nil
	}

	// Read state file
	data, err := os.ReadFile(config.StateFile)
	if err != nil {
		return err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	config.Projects = state.Projects
	return nil
}

// saveProjects saves projects to the state file
func saveProjects(config *Config) error {
	state := State{
		Projects: config.Projects,
		Version:  "1.0",
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(config.StateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(config.StateFile, data, 0644)
}
