package response

import "time"

type PullRequest struct {
	ID                int64      `json:"id"`
	NodeID            string     `json:"node_id"`
	Number            int        `json:"number"`
	Title             string     `json:"title"`
	Body              *string    `json:"body"`
	User              User       `json:"user"`
	Labels            []Label    `json:"labels"`
	State             string     `json:"state"`
	Locked            bool       `json:"locked"`
	Assignee          *User      `json:"assignee"`
	Assignees         []User     `json:"assignees"`
	RequestedReviewers []User    `json:"requested_reviewers"`
	RequestedTeams    []Team     `json:"requested_teams"`
	Head              PullRequestBranch `json:"head"`
	Base              PullRequestBranch `json:"base"`
	AuthorAssociation string     `json:"author_association"`
	AutoMerge         *AutoMerge `json:"auto_merge"`
	Draft             bool       `json:"draft"`
	Merged            bool       `json:"merged"`
	Mergeable         *bool      `json:"mergeable"`
	Rebaseable        *bool      `json:"rebaseable"`
	MergeableState    string     `json:"mergeable_state"`
	MergedBy          *User      `json:"merged_by"`
	Comments          int        `json:"comments"`
	ReviewComments    int        `json:"review_comments"`
	MaintainerCanModify bool     `json:"maintainer_can_modify"`
	Commits           int        `json:"commits"`
	Additions         int        `json:"additions"`
	Deletions         int        `json:"deletions"`
	ChangedFiles      int        `json:"changed_files"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	ClosedAt          *time.Time `json:"closed_at"`
	MergedAt          *time.Time `json:"merged_at"`
	HTMLURL           string     `json:"html_url"`
	URL               string     `json:"url"`
	IssueURL          string     `json:"issue_url"`
	CommitsURL        string     `json:"commits_url"`
	ReviewCommentsURL string     `json:"review_comments_url"`
	ReviewCommentURL  string     `json:"review_comment_url"`
	CommentsURL       string     `json:"comments_url"`
	StatusesURL       string     `json:"statuses_url"`
	DiffURL           string     `json:"diff_url"`
	PatchURL          string     `json:"patch_url"`
}

type PullRequestBranch struct {
	Label string      `json:"label"`
	Ref   string      `json:"ref"`
	SHA   string      `json:"sha"`
	User  User        `json:"user"`
	Repo  *Repository `json:"repo"`
}

type Team struct {
	ID              int64  `json:"id"`
	NodeID          string `json:"node_id"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	Description     *string `json:"description"`
	Privacy         string `json:"privacy"`
	Permission      string `json:"permission"`
	URL             string `json:"url"`
	HTMLURL         string `json:"html_url"`
	MembersURL      string `json:"members_url"`
	RepositoriesURL string `json:"repositories_url"`
	Parent          *Team  `json:"parent"`
}

type AutoMerge struct {
	EnabledBy     User   `json:"enabled_by"`
	MergeMethod   string `json:"merge_method"`
	CommitTitle   string `json:"commit_title"`
	CommitMessage string `json:"commit_message"`
}

// CreatePullRequestRequest represents the request to create a GitHub pull request
type CreatePullRequestRequest struct {
	Title               string   `json:"title"`
	Head                string   `json:"head"`
	Base                string   `json:"base"`
	Body                *string  `json:"body,omitempty"`
	MaintainerCanModify *bool    `json:"maintainer_can_modify,omitempty"`
	Draft               *bool    `json:"draft,omitempty"`
	Issue               *int     `json:"issue,omitempty"`
}

// UpdatePullRequestRequest represents the request to update a GitHub pull request
type UpdatePullRequestRequest struct {
	Title               *string `json:"title,omitempty"`
	Body                *string `json:"body,omitempty"`
	State               *string `json:"state,omitempty"`
	Base                *string `json:"base,omitempty"`
	MaintainerCanModify *bool   `json:"maintainer_can_modify,omitempty"`
}

// RequestReviewersRequest represents the request to add reviewers to a pull request
type RequestReviewersRequest struct {
	Reviewers     []string `json:"reviewers,omitempty"`
	TeamReviewers []string `json:"team_reviewers,omitempty"`
}