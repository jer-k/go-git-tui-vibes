package ui

import (
	"strings"

	"go-git-tui/internal/git"

	"github.com/charmbracelet/lipgloss"
)

type CommitCarousel struct {
	commits     []git.CommitInfo
	cursor      int
	width       int
	height      int
	maxVisible  int
	selectedIdx int
	loading     bool
}

func NewCommitCarousel() *CommitCarousel {
	return &CommitCarousel{
		commits:     make([]git.CommitInfo, 0),
		cursor:      0,
		selectedIdx: 0,
		loading:     false,
	}
}

func (c *CommitCarousel) UpdateSize(width, height int) {
	c.width = width
	c.height = height

	sampleBox := NewCommitBox(git.CommitInfo{}, false)
	boxWidth := sampleBox.GetWidth()

	minSpacing := 15
	availableWidth := width - minSpacing
	c.maxVisible = availableWidth / (boxWidth + minSpacing)
	if c.maxVisible < 1 {
		c.maxVisible = 1
	}

	maxCursor := len(c.commits) - c.maxVisible
	if maxCursor < 0 {
		maxCursor = 0
	}
	if c.cursor > maxCursor {
		c.cursor = maxCursor
	}
}

func (c *CommitCarousel) ScrollLeft() {
	if c.cursor > 0 {
		c.cursor--
		if c.selectedIdx > c.cursor {
			c.selectedIdx = c.cursor
		}
	}
}

func (c *CommitCarousel) ScrollRight() {
	maxCursor := len(c.commits) - c.maxVisible
	if maxCursor < 0 {
		maxCursor = 0
	}
	if c.cursor < maxCursor {
		c.cursor++
		if c.selectedIdx < c.cursor {
			c.selectedIdx = c.cursor
		}
	}
}

func (c *CommitCarousel) NearEnd() bool {
	return c.cursor+c.maxVisible >= len(c.commits)-3
}

func (c *CommitCarousel) SetLoading(loading bool) {
	c.loading = loading
}

func (c *CommitCarousel) AppendCommits(commits []git.CommitInfo) {
	c.commits = append(c.commits, commits...)
}

func (c *CommitCarousel) SetCommits(commits []git.CommitInfo) {
	c.commits = commits
	c.cursor = 0
	c.selectedIdx = 0
}

func (c *CommitCarousel) GetCommits() []git.CommitInfo {
	return c.commits
}

func (c *CommitCarousel) visibleCommits() []git.CommitInfo {
	if len(c.commits) == 0 {
		return nil
	}

	end := c.cursor + c.maxVisible
	if end > len(c.commits) {
		end = len(c.commits)
	}
	return c.commits[c.cursor:end]
}

func (c *CommitCarousel) Render() string {
	if !c.loading && len(c.commits) == 0 {
		return lipgloss.NewStyle().
			Width(c.width).
			Align(lipgloss.Center).
			Render("No commits to display")
	}

	visible := c.visibleCommits()
	if len(visible) == 0 {
		if c.loading {
			return lipgloss.NewStyle().
				Width(c.width).
				Align(lipgloss.Center).
				Render("Loading commits...")
		}
		return ""
	}

	sampleBox := NewCommitBox(git.CommitInfo{}, false)
	boxWidth := sampleBox.GetWidth()
	dotLine := sampleBox.GetDotLine()

	totalBoxesWidth := boxWidth * len(visible)
	remainingSpace := c.width - totalBoxesWidth
	spacing := remainingSpace / (len(visible) + 1)
	if spacing < 15 {
		spacing = 15
	}
	lineWidth := spacing - 4

	boxes := make([][]string, len(visible))
	for i, commit := range visible {
		isSelected := (c.cursor + i) == c.selectedIdx
		box := NewCommitBox(commit, isSelected)
		boxes[i] = strings.Split(box.Render(), "\n")
	}

	height := len(boxes[0])
	var result []string

	for row := 0; row < height; row++ {
		var rowElements []string
		for i := 0; i < len(visible); i++ {
			rowElements = append(rowElements, boxes[i][row])

			if i < len(visible)-1 {
				if row == dotLine+1 {
					line := NewGradientLine(visible[i].Color, visible[i+1].Color, lineWidth)
					padding := strings.Repeat(" ", 2)
					rowElements = append(rowElements, padding+line.Render()+padding)
				} else {
					rowElements = append(rowElements, strings.Repeat(" ", lineWidth+4))
				}
			}
		}
		result = append(result, lipgloss.JoinHorizontal(lipgloss.Top, rowElements...))
	}

	content := lipgloss.JoinVertical(lipgloss.Top, result...)

	return lipgloss.NewStyle().
		Width(c.width).
		Align(lipgloss.Center).
		Render(content)
}
