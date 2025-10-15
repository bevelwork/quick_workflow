package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	qc "github.com/bevelwork/quick_color"
)

// AuthConfig represents stored authentication configuration
type AuthConfig struct {
	GitHubToken string `json:"github_token,omitempty"`
	GitLabToken string `json:"gitlab_token,omitempty"`
	GitLabHost  string `json:"gitlab_host,omitempty"`
}


// loginGitHub initiates GitHub authentication
func loginGitHub() error {
	fmt.Printf("%s\n", qc.Colorize("GitHub Authentication", qc.ColorBlue))
	fmt.Println()

	fmt.Printf("%s\n", qc.Colorize("To authenticate with GitHub:", qc.ColorYellow))
	fmt.Println("1. Go to https://github.com/settings/tokens")
	fmt.Println("2. Click 'Generate new token (classic)'")
	fmt.Println("3. Select scopes: repo, read:org, read:user, read:packages")
	fmt.Println("4. Copy the generated token")
	fmt.Println()

	// For now, we'll use a simple token input approach
	// In a real implementation, you'd want to use OAuth device flow
	fmt.Printf("%s Enter your GitHub Personal Access Token: ", qc.Colorize("Token:", qc.ColorYellow))
	
	var token string
	fmt.Scanln(&token)
	
	if token == "" {
		return fmt.Errorf("no token provided")
	}

	// Test the token by making a simple API call
	if err := testGitHubToken(token); err != nil {
		return fmt.Errorf("invalid token: %v", err)
	}

	// Save token
	if err := saveAuthConfig(AuthConfig{GitHubToken: token}); err != nil {
		return fmt.Errorf("failed to save authentication: %v", err)
	}

	fmt.Printf("%s Successfully authenticated with GitHub!\n", qc.Colorize("Success:", qc.ColorGreen))
	return nil
}

// loginGitLab initiates GitLab authentication
func loginGitLab(host string) error {
	if host == "" {
		host = "gitlab.com"
	}

	fmt.Printf("%s\n", qc.Colorize("GitLab Authentication", qc.ColorBlue))
	fmt.Printf("Host: %s\n", qc.ColorizeBold(host, qc.ColorCyan))
	fmt.Println()

	fmt.Printf("%s\n", qc.Colorize("To authenticate with GitLab:", qc.ColorYellow))
	fmt.Printf("1. Go to https://%s/-/profile/personal_access_tokens\n", host)
	fmt.Println("2. Click 'Add new token'")
	fmt.Println("3. Select scopes: api, read_repository")
	fmt.Println("4. Copy the generated token")
	fmt.Println()

	fmt.Printf("%s Enter your GitLab Personal Access Token: ", qc.Colorize("Token:", qc.ColorYellow))
	
	var token string
	fmt.Scanln(&token)
	
	if token == "" {
		return fmt.Errorf("no token provided")
	}

	// Test the token by making a simple API call
	if err := testGitLabToken(host, token); err != nil {
		return fmt.Errorf("invalid token: %v", err)
	}

	// Save token
	if err := saveAuthConfig(AuthConfig{GitLabToken: token, GitLabHost: host}); err != nil {
		return fmt.Errorf("failed to save authentication: %v", err)
	}

	fmt.Printf("%s Successfully authenticated with GitLab (%s)!\n", qc.Colorize("Success:", qc.ColorGreen), host)
	return nil
}

// testGitHubToken tests a GitHub token by making a simple API call
func testGitHubToken(token string) error {
	client := &http.Client{Timeout: 30 * time.Second}
	
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	return nil
}

// testGitLabToken tests a GitLab token by making a simple API call
func testGitLabToken(host, token string) error {
	client := &http.Client{Timeout: 30 * time.Second}
	
	baseURL := fmt.Sprintf("https://%s", host)
	if !strings.HasPrefix(host, "http") {
		baseURL = fmt.Sprintf("https://%s", host)
	}
	
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v4/user", baseURL), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitLab API returned status %d", resp.StatusCode)
	}

	return nil
}


// saveAuthConfig saves authentication configuration to file
func saveAuthConfig(config AuthConfig) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	authDir := fmt.Sprintf("%s/.config/quick_workflow", homeDir)
	if err := os.MkdirAll(authDir, 0755); err != nil {
		return err
	}

	authFile := fmt.Sprintf("%s/auth.json", authDir)
	
	// Load existing config if it exists
	existingConfig := AuthConfig{}
	if data, err := os.ReadFile(authFile); err == nil {
		json.Unmarshal(data, &existingConfig)
	}

	// Merge with existing config
	if config.GitHubToken != "" {
		existingConfig.GitHubToken = config.GitHubToken
	}
	if config.GitLabToken != "" {
		existingConfig.GitLabToken = config.GitLabToken
	}
	if config.GitLabHost != "" {
		existingConfig.GitLabHost = config.GitLabHost
	}

	data, err := json.MarshalIndent(existingConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(authFile, data, 0600)
}

// loadAuthConfig loads authentication configuration from file
func loadAuthConfig() (*AuthConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	authFile := fmt.Sprintf("%s/.config/quick_workflow/auth.json", homeDir)
	data, err := os.ReadFile(authFile)
	if err != nil {
		return nil, err
	}

	var config AuthConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// showAuthStatus displays current authentication status
func showAuthStatus() {
	config, err := loadAuthConfig()
	if err != nil {
		fmt.Printf("%s No authentication found\n", qc.Colorize("Info:", qc.ColorCyan))
		return
	}

	fmt.Printf("%s\n", qc.Colorize("Authentication Status:", qc.ColorBlue))
	
	if config.GitHubToken != "" {
		fmt.Printf("GitHub: %s\n", qc.Colorize("✓ Authenticated", qc.ColorGreen))
	} else {
		fmt.Printf("GitHub: %s\n", qc.Colorize("✗ Not authenticated", qc.ColorRed))
	}

	if config.GitLabToken != "" {
		host := config.GitLabHost
		if host == "" {
			host = "gitlab.com"
		}
		fmt.Printf("GitLab (%s): %s\n", host, qc.Colorize("✓ Authenticated", qc.ColorGreen))
	} else {
		fmt.Printf("GitLab: %s\n", qc.Colorize("✗ Not authenticated", qc.ColorRed))
	}
}

// logout removes authentication tokens
func logout(platform string) error {
	config, err := loadAuthConfig()
	if err != nil {
		return fmt.Errorf("no authentication found")
	}

	switch platform {
	case "github":
		config.GitHubToken = ""
	case "gitlab":
		config.GitLabToken = ""
		config.GitLabHost = ""
	case "all":
		config.GitHubToken = ""
		config.GitLabToken = ""
		config.GitLabHost = ""
	default:
		return fmt.Errorf("invalid platform: %s", platform)
	}

	return saveAuthConfig(*config)
}
