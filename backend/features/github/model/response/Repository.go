package response

import "time"

type Repository struct {
	ID           int64     `json:"id"`
	NodeID       string    `json:"node_id"`
	Name         string    `json:"name"`
	FullName     string    `json:"full_name"`
	Owner        Owner     `json:"owner"`
	Private      bool      `json:"private"`
	HTMLURL      string    `json:"html_url"`
	Description  *string   `json:"description"`
	Fork         bool      `json:"fork"`
	URL          string    `json:"url"`
	ArchiveURL   string    `json:"archive_url"`
	AssigneesURL string    `json:"assignees_url"`
	BlobsURL     string    `json:"blobs_url"`
	BranchesURL  string    `json:"branches_url"`
	CloneURL     string    `json:"clone_url"`
	SSHURL       string    `json:"ssh_url"`
	GitURL       string    `json:"git_url"`
	DefaultBranch string   `json:"default_branch"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	PushedAt     *time.Time `json:"pushed_at"`
	Size         int       `json:"size"`
	StargazersCount int    `json:"stargazers_count"`
	WatchersCount   int    `json:"watchers_count"`
	Language        *string `json:"language"`
	ForksCount      int    `json:"forks_count"`
	Archived        bool   `json:"archived"`
	Disabled        bool   `json:"disabled"`
	OpenIssuesCount int    `json:"open_issues_count"`
	License         *License `json:"license"`
	Topics          []string `json:"topics"`
	Visibility      string   `json:"visibility"`
}

type Owner struct {
	Login             string `json:"login"`
	ID                int64  `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}

type License struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	SPDXID string `json:"spdx_id"`
	URL    *string `json:"url"`
	NodeID string `json:"node_id"`
}

// LocalRepository represents the status of a repository in local storage
type LocalRepository struct {
	Name         string    `json:"name"`
	FullName     string    `json:"full_name"`
	LocalPath    string    `json:"local_path"`
	CloneURL     string    `json:"clone_url"`
	DefaultBranch string   `json:"default_branch"`
	LastUpdated  time.Time `json:"last_updated"`
	IsCloned     bool      `json:"is_cloned"`
	LastCommit   string    `json:"last_commit"`
	Status       string    `json:"status"` // "up-to-date", "behind", "ahead", "diverged", "error"
}