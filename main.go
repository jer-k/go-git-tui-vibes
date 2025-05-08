package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"go-git-tui/internal/git"
	"go-git-tui/internal/model"
)

func initialModel() (*model.Model, error) {
	if len(os.Args) < 2 {
		return nil, fmt.Errorf("Please provide a path to a git repository\nUsage: go-git-tui <path-to-repo>")
	}

	repoPath := os.Args[1]
	m, err := model.New(repoPath)
	if err != nil {
		return nil, fmt.Errorf("Error initializing model: %v", err)
	}

	// Initialize git repository
	_, err = git.New(repoPath)
	if err != nil {
		return nil, fmt.Errorf("Error opening repository: %v", err)
	}

	return m, nil
}

func main() {
	m, err := initialModel()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}