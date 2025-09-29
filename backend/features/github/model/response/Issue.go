package response

import "time"

type Issue struct {
	ID          int64     `json:"id"`
	NodeID      string    `json:"node_id"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Body        *string   `json:"body"`
	User        User      `json:"user"`
	Labels      []Label   `json:"labels"`
	State       string    `json:"state"`
	Locked      bool      `json:"locked"`
	Assignee    *User     `json:"assignee"`
	Assignees   []User    `json:"assignees"`
	Milestone   *Milestone `json:"milestone"`
	Comments    int       `json:"comments"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ClosedAt    *time.Time `json:"closed_at"`
	AuthorAssociation string `json:"author_association"`
	ActiveLockReason  *string `json:"active_lock_reason"`
	HTMLURL     string    `json:"html_url"`
	URL         string    `json:"url"`
	CommentsURL string    `json:"comments_url"`
	EventsURL   string    `json:"events_url"`
	LabelsURL   string    `json:"labels_url"`
	RepositoryURL string  `json:"repository_url"`
}

type Label struct {
	ID          int64   `json:"id"`
	NodeID      string  `json:"node_id"`
	URL         string  `json:"url"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Color       string  `json:"color"`
	Default     bool    `json:"default"`
}

type Milestone struct {
	ID           int64      `json:"id"`
	NodeID       string     `json:"node_id"`
	Number       int        `json:"number"`
	Title        string     `json:"title"`
	Description  *string    `json:"description"`
	Creator      User       `json:"creator"`
	OpenIssues   int        `json:"open_issues"`
	ClosedIssues int        `json:"closed_issues"`
	State        string     `json:"state"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DueOn        *time.Time `json:"due_on"`
	ClosedAt     *time.Time `json:"closed_at"`
	HTMLURL      string     `json:"html_url"`
	LabelsURL    string     `json:"labels_url"`
	URL          string     `json:"url"`
}

// CreateIssueRequest represents the request to create a GitHub issue
type CreateIssueRequest struct {
	Title     string   `json:"title"`
	Body      *string  `json:"body,omitempty"`
	Assignee  *string  `json:"assignee,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
	Milestone *int     `json:"milestone,omitempty"`
	Labels    []string `json:"labels,omitempty"`
}

// UpdateIssueRequest represents the request to update a GitHub issue
type UpdateIssueRequest struct {
	Title     *string  `json:"title,omitempty"`
	Body      *string  `json:"body,omitempty"`
	State     *string  `json:"state,omitempty"`
	Assignee  *string  `json:"assignee,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
	Milestone *int     `json:"milestone,omitempty"`
	Labels    []string `json:"labels,omitempty"`
}