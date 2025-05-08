package git

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const (
	defaultPageSize = 20
	maxCachedPages  = 5
)

type CommitInfo struct {
	Hash         string
	ShortHash    string
	Author       string
	AuthorEmail  string
	Committer    string
	CommitEmail  string
	Message      string
	FullMessage  string
	Date         time.Time
	Color        string
	ParentHashes []string
	Branches     []string
	Tags         []string
	Stats        struct {
		Additions int
		Deletions int
		Files     int
	}
}

type Repository struct {
	repo     *git.Repository
	cache    *commitCache
	head     *plumbing.Reference
	branches map[string]*plumbing.Reference
	tags     map[string]*plumbing.Reference
	mu       sync.RWMutex
}

type commitCache struct {
	commits  map[string]*CommitInfo
	pages    [][]string // Stores hash sequences for each page
	pageSize int
	mu       sync.RWMutex
}

func newCommitCache(pageSize int) *commitCache {
	return &commitCache{
		commits:  make(map[string]*CommitInfo),
		pages:    make([][]string, 0),
		pageSize: pageSize,
	}
}

func New(path string) (*Repository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	r := &Repository{
		repo:     repo,
		cache:    newCommitCache(defaultPageSize),
		branches: make(map[string]*plumbing.Reference),
		tags:     make(map[string]*plumbing.Reference),
	}

	if err := r.updateRefs(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Repository) updateRefs() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Update HEAD
	head, err := r.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}
	r.head = head

	// Update branches
	branches, err := r.repo.Branches()
	if err != nil {
		return fmt.Errorf("failed to get branches: %w", err)
	}
	branches.ForEach(func(ref *plumbing.Reference) error {
		r.branches[ref.Name().Short()] = ref
		return nil
	})

	// Update tags
	tags, err := r.repo.Tags()
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}
	tags.ForEach(func(ref *plumbing.Reference) error {
		r.tags[ref.Name().Short()] = ref
		return nil
	})

	return nil
}

func (r *Repository) LoadCommits(startHash string, limit int) ([]CommitInfo, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var start plumbing.Hash
	if startHash != "" {
		start = plumbing.NewHash(startHash)
	} else {
		start = r.head.Hash()
	}

	commits := make([]CommitInfo, 0, limit)
	hasMore := false

	cIter, err := r.repo.Log(&git.LogOptions{
		From:  start,
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to get commit iterator: %w", err)
	}

	seen := make(map[string]bool)
	count := 0

	err = cIter.ForEach(func(c *object.Commit) error {
		if count >= limit {
			hasMore = true
			return fmt.Errorf("limit reached")
		}

		hash := c.Hash.String()
		if seen[hash] {
			return nil
		}
		seen[hash] = true

		// Get stats
		stats, _ := c.Stats()
		additions, deletions := 0, 0
		for _, stat := range stats {
			additions += stat.Addition
			deletions += stat.Deletion
		}

		// Get parent hashes
		parentHashes := make([]string, len(c.ParentHashes))
		for i, hash := range c.ParentHashes {
			parentHashes[i] = hash.String()[:7]
		}

		// Create commit info
		commit := CommitInfo{
			Hash:         hash,
			ShortHash:    hash[:7],
			Author:       c.Author.Name,
			AuthorEmail:  c.Author.Email,
			Committer:    c.Committer.Name,
			CommitEmail:  c.Committer.Email,
			Message:      c.Message,
			FullMessage:  c.Message,
			Date:         c.Author.When,
			Color:        "#" + hash[:6],
			ParentHashes: parentHashes,
			Stats: struct {
				Additions int
				Deletions int
				Files     int
			}{
				Additions: additions,
				Deletions: deletions,
				Files:     len(stats),
			},
		}

		// Add branch and tag information
		for name, ref := range r.branches {
			if ref.Hash() == c.Hash {
				commit.Branches = append(commit.Branches, name)
			}
		}
		for name, ref := range r.tags {
			if ref.Hash() == c.Hash {
				commit.Tags = append(commit.Tags, name)
			}
		}

		commits = append(commits, commit)
		count++
		return nil
	})

	if err != nil && err.Error() != "limit reached" {
		return nil, false, fmt.Errorf("failed to iterate commits: %w", err)
	}

	// Cache the results
	r.cache.mu.Lock()
	for _, commit := range commits {
		commitCopy := commit
		r.cache.commits[commit.Hash] = &commitCopy
	}
	r.cache.mu.Unlock()

	return commits, hasMore, nil
}

func (r *Repository) GetCommitByHash(hash string) (*CommitInfo, error) {
	// Check cache first
	r.cache.mu.RLock()
	if commit, ok := r.cache.commits[hash]; ok {
		r.cache.mu.RUnlock()
		return commit, nil
	}
	r.cache.mu.RUnlock()

	// Not in cache, load it
	commits, _, err := r.LoadCommits(hash, 1)
	if err != nil {
		return nil, err
	}
	if len(commits) == 0 {
		return nil, fmt.Errorf("commit not found")
	}

	return &commits[0], nil
}

func (r *Repository) GetCurrentBranch() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.head == nil {
		return ""
	}
	return r.head.Name().Short()
}

func (r *Repository) GetBranches() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	branches := make([]string, 0, len(r.branches))
	for name := range r.branches {
		branches = append(branches, name)
	}
	return branches
}

func (r *Repository) GetTags() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tags := make([]string, 0, len(r.tags))
	for name := range r.tags {
		tags = append(tags, name)
	}
	return tags
}

func (r *Repository) Refresh() error {
	return r.updateRefs()
}