package model

import (
	"go-git-tui/internal/git"
	"go-git-tui/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

type loadCommitsMsg struct {
	commits []git.CommitInfo
	hasMore bool
	err     error
}

type Model struct {
	repo     *git.Repository
	carousel *ui.CommitCarousel
	width    int
	height   int
	ready    bool
	err      error
	loading  bool
	lastHash string
	hasMore  bool
}

func New(repoPath string) (*Model, error) {
	repo, err := git.New(repoPath)
	if err != nil {
		return nil, err
	}

	m := &Model{
		repo:     repo,
		carousel: ui.NewCommitCarousel(),
		hasMore:  true,
	}

	return m, nil
}

func (m *Model) loadMoreCommits() tea.Cmd {
	return func() tea.Msg {
		if !m.hasMore || m.loading {
			return nil
		}

		m.loading = true
		commits, hasMore, err := m.repo.LoadCommits(m.lastHash, 20)
		if err != nil {
			return loadCommitsMsg{err: err}
		}

		if len(commits) > 0 {
			m.lastHash = commits[len(commits)-1].Hash
		}

		return loadCommitsMsg{
			commits: commits,
			hasMore: hasMore,
			err:     nil,
		}
	}
}

func (m *Model) Init() tea.Cmd {
	return m.loadMoreCommits()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case loadCommitsMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}

		m.hasMore = msg.hasMore
		if len(msg.commits) > 0 {
			m.carousel.AppendCommits(msg.commits)

			if m.carousel.NearEnd() && m.hasMore {
				cmds = append(cmds, m.loadMoreCommits())
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "left":
			m.carousel.ScrollLeft()
		case "right":
			m.carousel.ScrollRight()
			if m.carousel.NearEnd() && m.hasMore && !m.loading {
				cmds = append(cmds, m.loadMoreCommits())
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.carousel.UpdateSize(msg.Width, msg.Height)
			m.ready = true

			if !m.loading && len(m.carousel.GetCommits()) == 0 {
				cmds = append(cmds, m.loadMoreCommits())
			}
		} else {
			m.carousel.UpdateSize(msg.Width, msg.Height)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if !m.ready {
		return "\nInitializing..."
	}

	if m.err != nil {
		return "Error: " + m.err.Error()
	}

	return m.carousel.Render()
}