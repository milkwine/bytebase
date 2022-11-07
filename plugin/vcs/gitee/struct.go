package gitee

import (
	"strconv"
	"time"

	"github.com/bytebase/bytebase/plugin/vcs"
)

type apiRespReadFile struct {
	Type        string `json:"type"`
	Encoding    string `json:"encoding"`
	Size        int    `json:"size"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Content     string `json:"content"`
	Sha         string `json:"sha"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	DownloadURL string `json:"download_url"`
	Links       struct {
		Self string `json:"self"`
		HTML string `json:"html"`
	} `json:"_links"`
}

func (r apiRespFetchUserRepo) toVCSRepo() *vcs.Repository {
	return &vcs.Repository{
		ID:       int64(r.ID),
		Name:     r.Name,
		FullPath: r.FullName,
		WebURL:   r.HTMLURL,
	}
}

type apiRespFetchUserRepo struct {
	ID        int    `json:"id"`
	FullName  string `json:"full_name"`
	HumanName string `json:"human_name"`
	URL       string `json:"url"`
	Namespace struct {
		ID      int    `json:"id"`
		Type    string `json:"type"`
		Name    string `json:"name"`
		Path    string `json:"path"`
		HTMLURL string `json:"html_url"`
	} `json:"namespace"`
	Path  string `json:"path"`
	Name  string `json:"name"`
	Owner struct {
		ID                int    `json:"id"`
		Login             string `json:"login"`
		Name              string `json:"name"`
		AvatarURL         string `json:"avatar_url"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		Remark            string `json:"remark"`
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
	} `json:"owner"`
	Assigner struct {
		ID                int    `json:"id"`
		Login             string `json:"login"`
		Name              string `json:"name"`
		AvatarURL         string `json:"avatar_url"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		Remark            string `json:"remark"`
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
	} `json:"assigner"`
	Description         string      `json:"description"`
	Private             bool        `json:"private"`
	Public              bool        `json:"public"`
	Internal            bool        `json:"internal"`
	Fork                bool        `json:"fork"`
	HTMLURL             string      `json:"html_url"`
	SSHURL              string      `json:"ssh_url"`
	ForksURL            string      `json:"forks_url"`
	KeysURL             string      `json:"keys_url"`
	CollaboratorsURL    string      `json:"collaborators_url"`
	HooksURL            string      `json:"hooks_url"`
	BranchesURL         string      `json:"branches_url"`
	TagsURL             string      `json:"tags_url"`
	BlobsURL            string      `json:"blobs_url"`
	StargazersURL       string      `json:"stargazers_url"`
	ContributorsURL     string      `json:"contributors_url"`
	CommitsURL          string      `json:"commits_url"`
	CommentsURL         string      `json:"comments_url"`
	IssueCommentURL     string      `json:"issue_comment_url"`
	IssuesURL           string      `json:"issues_url"`
	PullsURL            string      `json:"pulls_url"`
	MilestonesURL       string      `json:"milestones_url"`
	NotificationsURL    string      `json:"notifications_url"`
	LabelsURL           string      `json:"labels_url"`
	ReleasesURL         string      `json:"releases_url"`
	Recommend           bool        `json:"recommend"`
	Gvp                 bool        `json:"gvp"`
	Homepage            interface{} `json:"homepage"`
	Language            interface{} `json:"language"`
	ForksCount          int         `json:"forks_count"`
	StargazersCount     int         `json:"stargazers_count"`
	WatchersCount       int         `json:"watchers_count"`
	DefaultBranch       interface{} `json:"default_branch"`
	OpenIssuesCount     int         `json:"open_issues_count"`
	HasIssues           bool        `json:"has_issues"`
	HasWiki             bool        `json:"has_wiki"`
	IssueComment        bool        `json:"issue_comment"`
	CanComment          bool        `json:"can_comment"`
	PullRequestsEnabled bool        `json:"pull_requests_enabled"`
	HasPage             bool        `json:"has_page"`
	License             interface{} `json:"license"`
	Outsourced          bool        `json:"outsourced"`
	ProjectCreator      string      `json:"project_creator"`
	Members             []string    `json:"members"`
	PushedAt            interface{} `json:"pushed_at"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at"`
	Parent              interface{} `json:"parent"`
	Paas                interface{} `json:"paas"`
	Stared              bool        `json:"stared"`
	Watched             bool        `json:"watched"`
	Permission          struct {
		Pull  bool `json:"pull"`
		Push  bool `json:"push"`
		Admin bool `json:"admin"`
	} `json:"permission"`
	Relation        string `json:"relation"`
	AssigneesNumber int    `json:"assignees_number"`
	TestersNumber   int    `json:"testers_number"`
	Assignee        []struct {
		ID                int    `json:"id"`
		Login             string `json:"login"`
		Name              string `json:"name"`
		AvatarURL         string `json:"avatar_url"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		Remark            string `json:"remark"`
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
	} `json:"assignee"`
	Testers []struct {
		ID                int    `json:"id"`
		Login             string `json:"login"`
		Name              string `json:"name"`
		AvatarURL         string `json:"avatar_url"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		Remark            string `json:"remark"`
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
	} `json:"testers"`
	Status        string        `json:"status"`
	Programs      []interface{} `json:"programs"`
	Enterprise    interface{}   `json:"enterprise"`
	ProjectLabels []interface{} `json:"project_labels"`
}

type apiReqCreateWebHook struct {
	URL                 string `json:"url"`
	EncryptionType      int    `json:"encryption_type"`
	Password            string `json:"password"`
	PushEvents          bool   `json:"push_events"`
	TagPushEvents       bool   `json:"tag_push_events"`
	IssuesEvents        bool   `json:"issues_events"`
	NoteEvents          bool   `json:"note_events"`
	MergeRequestsEvents bool   `json:"merge_requests_events"`
}

type apiRespCreateWebHook struct {
	ID                  int         `json:"id"`
	URL                 string      `json:"url"`
	CreatedAt           time.Time   `json:"created_at"`
	Password            string      `json:"password"`
	ProjectID           int         `json:"project_id"`
	Result              string      `json:"result"`
	ResultCode          interface{} `json:"result_code"`
	PushEvents          bool        `json:"push_events"`
	TagPushEvents       bool        `json:"tag_push_events"`
	IssuesEvents        bool        `json:"issues_events"`
	NoteEvents          bool        `json:"note_events"`
	MergeRequestsEvents bool        `json:"merge_requests_events"`
}

func (e giteePushEvent) toVcsPushEvent() vcs.PushEvent {
	commits := make([]vcs.Commit, 0)
	for _, c := range e.Commits {
		commits = append(commits, c.toVcsCommit())
	}

	ret := vcs.PushEvent{
		VCSType:            vcs.GiteeCom,
		BaseDirectory:      "",
		Ref:                e.Ref,
		RepositoryID:       strconv.Itoa(e.Repository.ID),
		RepositoryURL:      e.Repository.HTMLURL,
		RepositoryFullPath: e.Repository.FullName,
		AuthorName:         e.UserName,
		CommitList:         commits,
	}

	return ret
}

func (c giteeCommit) toVcsCommit() vcs.Commit {
	return vcs.Commit{
		ID:           c.ID,
		Title:        c.Message, // have no title here
		Message:      c.Message,
		CreatedTs:    c.Timestamp.Unix(),
		URL:          c.URL,
		AuthorName:   c.Author.Name,
		AuthorEmail:  c.Author.Email,
		AddedList:    c.Added,
		ModifiedList: c.Modified,
	}
}

type giteeCommit struct {
	ID        string    `json:"id"`
	TreeID    string    `json:"tree_id"`
	ParentIds []string  `json:"parent_ids"`
	Distinct  bool      `json:"distinct"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	URL       string    `json:"url"`
	Author    struct {
		Time     time.Time `json:"time"`
		ID       int       `json:"id"`
		Name     string    `json:"name"`
		Email    string    `json:"email"`
		Username string    `json:"username"`
		UserName string    `json:"user_name"`
		URL      string    `json:"url"`
	} `json:"author"`
	Committer struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
		UserName string `json:"user_name"`
		URL      string `json:"url"`
	} `json:"committer"`
	Added    []string `json:"added"`
	Removed  []string `json:"removed"`
	Modified []string `json:"modified"`
}

type giteePushEvent struct {
	Ref                string        `json:"ref"`
	Before             string        `json:"before"`
	After              string        `json:"after"`
	Created            bool          `json:"created"`
	Deleted            bool          `json:"deleted"`
	Compare            string        `json:"compare"`
	Commits            []giteeCommit `json:"commits"`
	HeadCommit         giteeCommit   `json:"head_commit"`
	TotalCommitsCount  int           `json:"total_commits_count"`
	CommitsMoreThanTen bool          `json:"commits_more_than_ten"`
	Repository         struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Path     string `json:"path"`
		FullName string `json:"full_name"`
		Owner    struct {
			Login     string `json:"login"`
			AvatarURL string `json:"avatar_url"`
			HTMLURL   string `json:"html_url"`
			Type      string `json:"type"`
			SiteAdmin bool   `json:"site_admin"`
			ID        int    `json:"id"`
			Name      string `json:"name"`
			Email     string `json:"email"`
			Username  string `json:"username"`
			UserName  string `json:"user_name"`
			URL       string `json:"url"`
		} `json:"owner"`
		Private           bool        `json:"private"`
		HTMLURL           string      `json:"html_url"`
		URL               string      `json:"url"`
		Description       string      `json:"description"`
		Fork              bool        `json:"fork"`
		CreatedAt         time.Time   `json:"created_at"`
		UpdatedAt         time.Time   `json:"updated_at"`
		PushedAt          time.Time   `json:"pushed_at"`
		GitURL            string      `json:"git_url"`
		SSHURL            string      `json:"ssh_url"`
		CloneURL          string      `json:"clone_url"`
		SvnURL            string      `json:"svn_url"`
		GitHTTPURL        string      `json:"git_http_url"`
		GitSSHURL         string      `json:"git_ssh_url"`
		GitSvnURL         string      `json:"git_svn_url"`
		Homepage          interface{} `json:"homepage"`
		StargazersCount   int         `json:"stargazers_count"`
		WatchersCount     int         `json:"watchers_count"`
		ForksCount        int         `json:"forks_count"`
		Language          string      `json:"language"`
		HasIssues         bool        `json:"has_issues"`
		HasWiki           bool        `json:"has_wiki"`
		HasPages          bool        `json:"has_pages"`
		License           interface{} `json:"license"`
		OpenIssuesCount   int         `json:"open_issues_count"`
		DefaultBranch     string      `json:"default_branch"`
		Namespace         string      `json:"namespace"`
		NameWithNamespace string      `json:"name_with_namespace"`
		PathWithNamespace string      `json:"path_with_namespace"`
	} `json:"repository"`
	Project struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Path     string `json:"path"`
		FullName string `json:"full_name"`
		Owner    struct {
			Login     string `json:"login"`
			AvatarURL string `json:"avatar_url"`
			HTMLURL   string `json:"html_url"`
			Type      string `json:"type"`
			SiteAdmin bool   `json:"site_admin"`
			ID        int    `json:"id"`
			Name      string `json:"name"`
			Email     string `json:"email"`
			Username  string `json:"username"`
			UserName  string `json:"user_name"`
			URL       string `json:"url"`
		} `json:"owner"`
		Private           bool        `json:"private"`
		HTMLURL           string      `json:"html_url"`
		URL               string      `json:"url"`
		Description       string      `json:"description"`
		Fork              bool        `json:"fork"`
		CreatedAt         time.Time   `json:"created_at"`
		UpdatedAt         time.Time   `json:"updated_at"`
		PushedAt          time.Time   `json:"pushed_at"`
		GitURL            string      `json:"git_url"`
		SSHURL            string      `json:"ssh_url"`
		CloneURL          string      `json:"clone_url"`
		SvnURL            string      `json:"svn_url"`
		GitHTTPURL        string      `json:"git_http_url"`
		GitSSHURL         string      `json:"git_ssh_url"`
		GitSvnURL         string      `json:"git_svn_url"`
		Homepage          interface{} `json:"homepage"`
		StargazersCount   int         `json:"stargazers_count"`
		WatchersCount     int         `json:"watchers_count"`
		ForksCount        int         `json:"forks_count"`
		Language          string      `json:"language"`
		HasIssues         bool        `json:"has_issues"`
		HasWiki           bool        `json:"has_wiki"`
		HasPages          bool        `json:"has_pages"`
		License           interface{} `json:"license"`
		OpenIssuesCount   int         `json:"open_issues_count"`
		DefaultBranch     string      `json:"default_branch"`
		Namespace         string      `json:"namespace"`
		NameWithNamespace string      `json:"name_with_namespace"`
		PathWithNamespace string      `json:"path_with_namespace"`
	} `json:"project"`
	UserID   int    `json:"user_id"`
	UserName string `json:"user_name"`
	User     struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
		UserName string `json:"user_name"`
		URL      string `json:"url"`
	} `json:"user"`
	Pusher struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
		UserName string `json:"user_name"`
		URL      string `json:"url"`
	} `json:"pusher"`
	Sender struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
		HTMLURL   string `json:"html_url"`
		Type      string `json:"type"`
		SiteAdmin bool   `json:"site_admin"`
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		Username  string `json:"username"`
		UserName  string `json:"user_name"`
		URL       string `json:"url"`
	} `json:"sender"`
	Enterprise struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"enterprise"`
	HookName  string `json:"hook_name"`
	HookID    int    `json:"hook_id"`
	HookURL   string `json:"hook_url"`
	Password  string `json:"password"`
	Timestamp string `json:"timestamp"`
	Sign      string `json:"sign"`
}

type apiRespUser struct {
	ID                int         `json:"id"`
	Login             string      `json:"login"`
	Name              string      `json:"name"`
	AvatarURL         string      `json:"avatar_url"`
	URL               string      `json:"url"`
	HTMLURL           string      `json:"html_url"`
	Remark            string      `json:"remark"`
	FollowersURL      string      `json:"followers_url"`
	FollowingURL      string      `json:"following_url"`
	GistsURL          string      `json:"gists_url"`
	StarredURL        string      `json:"starred_url"`
	SubscriptionsURL  string      `json:"subscriptions_url"`
	OrganizationsURL  string      `json:"organizations_url"`
	ReposURL          string      `json:"repos_url"`
	EventsURL         string      `json:"events_url"`
	ReceivedEventsURL string      `json:"received_events_url"`
	Type              string      `json:"type"`
	Blog              interface{} `json:"blog"`
	Weibo             interface{} `json:"weibo"`
	Bio               interface{} `json:"bio"`
	PublicRepos       int         `json:"public_repos"`
	PublicGists       int         `json:"public_gists"`
	Followers         int         `json:"followers"`
	Following         int         `json:"following"`
	Stared            int         `json:"stared"`
	Watched           int         `json:"watched"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	Email             string      `json:"email"`
}

type apiRespEmail struct {
	Email string   `json:"email"`
	State string   `json:"state"`
	Scope []string `json:"scope"`
}

type apiRespRepoCommit struct {
	URL         string `json:"url"`
	Sha         string `json:"sha"`
	HTMLURL     string `json:"html_url"`
	CommentsURL string `json:"comments_url"`
	Commit      struct {
		Author struct {
			Name  string    `json:"name"`
			Date  time.Time `json:"date"`
			Email string    `json:"email"`
		} `json:"author"`
		Committer struct {
			Name  string    `json:"name"`
			Date  time.Time `json:"date"`
			Email string    `json:"email"`
		} `json:"committer"`
		Message string `json:"message"`
		Tree    struct {
			Sha string `json:"sha"`
			URL string `json:"url"`
		} `json:"tree"`
	} `json:"commit"`
	Author struct {
		ID                int    `json:"id"`
		Login             string `json:"login"`
		Name              string `json:"name"`
		AvatarURL         string `json:"avatar_url"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		Remark            string `json:"remark"`
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
	} `json:"author"`
	Committer struct {
		ID                int    `json:"id"`
		Login             string `json:"login"`
		Name              string `json:"name"`
		AvatarURL         string `json:"avatar_url"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		Remark            string `json:"remark"`
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
	} `json:"committer"`
	Parents []struct {
		Sha string `json:"sha"`
		URL string `json:"url"`
	} `json:"parents"`
	Stats struct {
		ID        string `json:"id"`
		Additions int    `json:"additions"`
		Deletions int    `json:"deletions"`
		Total     int    `json:"total"`
	} `json:"stats"`
	Files []struct {
		Sha        string `json:"sha"`
		Filename   string `json:"filename"`
		Status     string `json:"status"`
		Additions  int    `json:"additions"`
		Deletions  int    `json:"deletions"`
		Changes    int    `json:"changes"`
		BlobURL    string `json:"blob_url"`
		RawURL     string `json:"raw_url"`
		ContentURL string `json:"content_url"`
		Patch      string `json:"patch"`
	} `json:"files"`
}

type apiRespRepoMember struct {
	ID                int    `json:"id"`
	Login             string `json:"login"`
	Name              string `json:"name"`
	AvatarURL         string `json:"avatar_url"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	Remark            string `json:"remark"`
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
	Permissions       struct {
		Pull  bool `json:"pull"`
		Push  bool `json:"push"`
		Admin bool `json:"admin"`
	} `json:"permissions"`
}

type apiRespPRFile struct {
	Sha       string `json:"sha"`
	Filename  string `json:"filename"`
	Status    string `json:"status"`
	Additions string `json:"additions"`
	Deletions string `json:"deletions"`
	BlobURL   string `json:"blob_url"`
	RawURL    string `json:"raw_url"`
	Patch     struct {
		Diff        string `json:"diff"`
		NewPath     string `json:"new_path"`
		OldPath     string `json:"old_path"`
		AMode       string `json:"a_mode"`
		BMode       string `json:"b_mode"`
		NewFile     bool   `json:"new_file"`
		RenamedFile bool   `json:"renamed_file"`
		DeletedFile bool   `json:"deleted_file"`
		TooLarge    bool   `json:"too_large"`
	} `json:"patch"`
}

type apiRespBranch struct {
	Name   string `json:"name"`
	Commit struct {
		Sha    string `json:"sha"`
		URL    string `json:"url"`
		Commit struct {
			Author struct {
				Name  string    `json:"name"`
				Date  time.Time `json:"date"`
				Email string    `json:"email"`
			} `json:"author"`
			URL     string `json:"url"`
			Message string `json:"message"`
			Tree    struct {
				Sha string `json:"sha"`
				URL string `json:"url"`
			} `json:"tree"`
			Committer struct {
				Name  string    `json:"name"`
				Date  time.Time `json:"date"`
				Email string    `json:"email"`
			} `json:"committer"`
		} `json:"commit"`
		Author struct {
			AvatarURL string `json:"avatar_url"`
			URL       string `json:"url"`
			ID        int    `json:"id"`
			Login     string `json:"login"`
		} `json:"author"`
		Parents []struct {
			Sha string `json:"sha"`
			URL string `json:"url"`
		} `json:"parents"`
		Committer struct {
			AvatarURL string `json:"avatar_url"`
			URL       string `json:"url"`
			ID        int    `json:"id"`
			Login     string `json:"login"`
		} `json:"committer"`
	} `json:"commit"`
	Links struct {
		HTML string `json:"html"`
		Self string `json:"self"`
	} `json:"_links"`
	Protected     bool   `json:"protected"`
	ProtectionURL string `json:"protection_url"`
}
