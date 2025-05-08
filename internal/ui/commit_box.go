package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"go-git-tui/internal/git"
)

const (
	boxWidth  = 40  // Fixed width for commit boxes
	boxHeight = 10  // Fixed height including both top and bottom borders
	dotLine   = 7   // Line where the dot appears (0-based)
)

var (
	boxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#909090")).
		Padding(0, 2)

	hashStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF"))

	branchStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2E8B57"))

	tagStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#DAA520")).
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1)

	authorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87CEEB"))

	dateStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	messageStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDDDDD"))

	statsStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#98FB98"))
)

type CommitBox struct {
	commit   git.CommitInfo
	selected bool
}

func NewCommitBox(commit git.CommitInfo, selected bool) CommitBox {
	return CommitBox{
		commit:   commit,
		selected: selected,
	}
}

func truncateText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}
	return text[:maxWidth-3] + "..."
}

func (c CommitBox) formatDate() string {
	return c.commit.Date.Format("2006-01-02 15:04")
}

func (c CommitBox) formatStats() string {
	return fmt.Sprintf("+%d -%d in %d files",
		c.commit.Stats.Additions,
		c.commit.Stats.Deletions,
		c.commit.Stats.Files)
}

func (c CommitBox) renderRefs() string {
	var refs []string

	for _, branch := range c.commit.Branches {
		refs = append(refs, branchStyle.Render(branch))
	}

	for _, tag := range c.commit.Tags {
		refs = append(refs, tagStyle.Render(tag))
	}

	if len(refs) == 0 {
		return " "
	}

	return strings.Join(refs, " ")
}

func (c CommitBox) GetWidth() int {
	return boxWidth
}

func (c CommitBox) GetHeight() int {
	return boxHeight
}

func (c CommitBox) GetDotLine() int {
	return dotLine
}

func (c CommitBox) GetDotPosition() (int, int) {
	x := boxWidth / 2
	y := dotLine + 1  // +1 to account for the top border
	return x, y
}

func (c CommitBox) Render() string {
	contentWidth := boxWidth - 4 // Account for padding and borders

	// Line 1: Hash
	hashLine := truncateText(c.commit.ShortHash, contentWidth)
	hashLine = hashStyle.Render(hashLine)

	// Line 2: Branch/Tags
	refsLine := truncateText(c.renderRefs(), contentWidth)

	// Line 3: Author
	authorLine := fmt.Sprintf("%s <%s>", c.commit.Author, c.commit.AuthorEmail)
	authorLine = truncateText(authorLine, contentWidth)
	authorLine = authorStyle.Render(authorLine)

	// Line 4: Date
	dateLine := c.formatDate()
	dateLine = truncateText(dateLine, contentWidth)
	dateLine = dateStyle.Render(dateLine)

	// Line 5: Stats
	statsLine := c.formatStats()
	statsLine = truncateText(statsLine, contentWidth)
	statsLine = statsStyle.Render(statsLine)

	// Line 6: Message (first line only)
	messageLine := strings.Split(c.commit.Message, "\n")[0]
	messageLine = truncateText(messageLine, contentWidth)
	messageLine = messageStyle.Render(messageLine)

	// Build lines array with fixed height
	lines := make([]string, boxHeight-2) // -2 for top and bottom borders
	lines[0] = hashLine
	lines[1] = refsLine
	lines[2] = authorLine
	lines[3] = dateLine
	lines[4] = statsLine
	lines[5] = messageLine
	lines[6] = ""
	lines[dotLine] = lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.commit.Color)).
			Render("●"))

	// Create the final box
	style := boxStyle.Copy()
	if c.selected {
		style = style.BorderForeground(lipgloss.Color("#00FF00"))
	}

	return style.
		Width(boxWidth).
		Height(boxHeight).
		Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}