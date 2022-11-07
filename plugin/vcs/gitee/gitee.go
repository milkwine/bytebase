// Package gitee is the plugin for GitHub.
package gitee

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/k0kubun/pp/v3"
	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/common"
	"github.com/bytebase/bytebase/plugin/vcs"
)

const (
	// apiPath is the API path.
	apiPath    = "api/v5"
	pageSize   = 100
	maxRetries = 3
)

var _ vcs.Provider = (*Provider)(nil)

func init() {
	vcs.Register(vcs.GiteeCom, newProvider)
}

// Provider is a GitLab self host VCS provider.
type Provider struct {
	client *http.Client
}

func newProvider(config vcs.ProviderConfig) vcs.Provider {
	if config.Client == nil {
		config.Client = &http.Client{}
	}
	return &Provider{config.Client}
}

// APIURL returns the API URL path of a GitLab instance.
func (*Provider) APIURL(instanceURL string) string {
	return fmt.Sprintf("%s/%s", instanceURL, apiPath)
}

// ExchangeOAuthToken exchanges OAuth content with the provided authorization code.
func (p *Provider) ExchangeOAuthToken(ctx context.Context, instanceURL string, oauthExchange *common.OAuthExchange) (*vcs.OAuthToken, error) {
	urlParams := &url.Values{}
	urlParams.Set("client_id", oauthExchange.ClientID)
	urlParams.Set("client_secret", oauthExchange.ClientSecret)
	urlParams.Set("code", oauthExchange.Code)
	urlParams.Set("redirect_uri", oauthExchange.RedirectURL)
	urlParams.Set("grant_type", "authorization_code")

	resp, err := oauthGrant(ctx, p.client, instanceURL, urlParams.Encode())
	if err != nil {
		return nil, err
	}

	return resp.toVCSOAuthToken(), nil
}

// TryLogin tries to fetch the user info from the current OAuth context.
func (p *Provider) TryLogin(ctx context.Context, oauthCtx common.OauthContext, instanceURL string) (*vcs.UserInfo, error) {
	i := p.newInstance(ctx, oauthCtx, instanceURL)

	// we can't get email here if user choose not to show their public email
	info, err := i.APILoginUser()
	if err != nil {
		return nil, err
	}

	emails, err := i.APILoginEmail()
	if err != nil {
		return nil, err
	}

	email := ""

find:
	for _, e := range emails {
		for _, scope := range e.Scope {
			if scope == "primary" {
				email = e.Email
				break find
			}
		}
	}

	// empty email will handle by the caller
	// TODO profileLink maybe should maintain by plugin itself
	return &vcs.UserInfo{
		PublicEmail: email, Name: info.Name,
		State: vcs.StateActive,
	}, nil
}

// FetchCommitByID fetches the commit data by its ID from the repository.
func (p *Provider) FetchCommitByID(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, repositoryID string, commitID string) (*vcs.Commit, error) {
	i := p.newInstance(ctx, oauthCtx, instanceURL)

	fullPath, err := i.findRepoFullPathByID(repositoryID)
	if err != nil {
		return nil, err
	}

	commit, err := i.APIGetCommit(fullPath, commitID)
	if err != nil {
		return nil, err
	}

	// some attr have no usage yet
	return &vcs.Commit{
		ID:         commit.Sha,
		CreatedTs:  commit.Commit.Author.Date.Unix(),
		AuthorName: commit.Commit.Author.Name,
	}, nil
}

// FetchUserInfo fetches user info of given user ID.
func (p *Provider) FetchUserInfo(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, user string) (*vcs.UserInfo, error) { //nolint:revive
	panic("not implemented(no caller)")
}

// FetchRepositoryActiveMemberList fetches all active members of a repository.
func (p *Provider) FetchRepositoryActiveMemberList(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, repositoryID string) ([]*vcs.RepositoryMember, error) { //nolint:revive
	// Gitee's api not provide member's email.
	return nil, errors.Errorf("Not support by Gitee")

	// i := p.newInstance(ctx, oauthCtx, instanceURL)

	// fullPath, err := i.findRepoFullPathByID(repositoryID)
	// if err != nil {
	//	return nil, err
	//}

	// members, err := i.APIRepoMember(fullPath)
	// if err != nil {
	//	return nil, err
	//}

	// ret := make([]*vcs.RepositoryMember, 0)
	// for _, member := range members {

	//	ret = append(ret, &vcs.RepositoryMember{
	//		Email:        "",
	//		Name:         member.Name,
	//		State:        vcs.StateActive,
	//		Role:         common.ProjectDeveloper,
	//		VCSRole:      "developer",
	//		RoleProvider: vcs.GiteeCom,
	//	})

	//}
	// return ret, nil
}

// FetchAllRepositoryList fetches all repositories where the authenticated user
// has a maintainer role, which is required to create webhook in the project.
func (p *Provider) FetchAllRepositoryList(ctx context.Context, oauthCtx common.OauthContext, instanceURL string) ([]*vcs.Repository, error) {
	repos, err := p.newInstance(ctx, oauthCtx, instanceURL).APIAllRepoList()

	if err != nil {
		return nil, err
	}
	ret := make([]*vcs.Repository, 0)

	for _, r := range repos {
		ret = append(ret, r.toVCSRepo())
	}

	return ret, err
}

// FetchRepositoryFileList fetches the all files from the given repository tree
// recursively.
func (p *Provider) FetchRepositoryFileList(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, repositoryID string, ref string, filePath string) ([]*vcs.RepositoryTreeNode, error) {
	i := p.newInstance(ctx, oauthCtx, instanceURL)
	fullPath, err := i.findRepoFullPathByID(repositoryID)
	if err != nil {
		return nil, err
	}

	files, err := i.APIListFile(fullPath, filePath, ref)
	if err != nil {
		return nil, err
	}
	ret := make([]*vcs.RepositoryTreeNode, 0)
	for _, f := range files {
		if f.Type == "file" {
			ret = append(ret, &vcs.RepositoryTreeNode{
				Path: f.Path,
				Type: f.Type,
			})
		}
	}
	return ret, nil
}

// CreateFile creates a file at given path in the repository.
func (p *Provider) CreateFile(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, repositoryID string, filePath string, fileCommit vcs.FileCommitCreate) error {
	pp.Println("create file", filePath)
	i := p.newInstance(ctx, oauthCtx, instanceURL)
	fullPath, err := i.findRepoFullPathByID(repositoryID)
	if err != nil {
		return err
	}

	form := struct {
		Content string `json:"content"`
		Message string `json:"message"`
		Branch  string `json:"branch"`
	}{
		base64.StdEncoding.EncodeToString([]byte(fileCommit.Content)),
		fileCommit.CommitMessage,
		fileCommit.Branch,
	}

	url := fmt.Sprintf("/api/v5/repos/%s/contents/%s", fullPath, url.QueryEscape(filePath))
	resp := make(map[string]interface{})
	s := i.s().Post(url).BodyJSON(&form)
	return i.req(s, &resp)
}

// OverwriteFile overwrites an existing file at given path in the repository.
func (p *Provider) OverwriteFile(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, repositoryID string, filePath string, fileCommit vcs.FileCommitCreate) error {
	pp.Println("put file", filePath)
	i := p.newInstance(ctx, oauthCtx, instanceURL)
	fullPath, err := i.findRepoFullPathByID(repositoryID)
	if err != nil {
		return err
	}

	form := struct {
		Content string `json:"content"`
		Sha     string `json:"sha"`
		Message string `json:"message"`
		Branch  string `json:"branch"`
	}{
		base64.StdEncoding.EncodeToString([]byte(fileCommit.Content)),
		fileCommit.LastCommitID,
		fileCommit.CommitMessage,
		fileCommit.Branch,
	}

	url := fmt.Sprintf("/api/v5/repos/%s/contents/%s", fullPath, url.QueryEscape(filePath))
	resp := make(map[string]interface{})
	s := i.s().Put(url).BodyJSON(&form)
	return i.req(s, &resp)
}

// ReadFileMeta reads the metadata of the given file in the repository.
func (p *Provider) ReadFileMeta(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, repositoryID string, filePath string, ref string) (*vcs.FileMeta, error) {
	pp.Println("read meta", filePath)
	i := p.newInstance(ctx, oauthCtx, instanceURL)
	fullPath, err := i.findRepoFullPathByID(repositoryID)
	if err != nil {
		return nil, err
	}

	file, err := i.APIReadFile(fullPath, filePath, ref)
	if err != nil {
		return nil, err
	}

	commit, err := i.APIFileLatestCommit(fullPath, filePath, ref)
	if err != nil {
		return nil, err
	}

	return &vcs.FileMeta{
		Name:         file.Name,
		Path:         file.Path,
		Size:         int64(file.Size),
		LastCommitID: commit.Sha,
	}, nil
}

// ReadFileContent reads the content of the given file in the repository.
func (p *Provider) ReadFileContent(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, repositoryID string, filePath string, ref string) (string, error) {
	i := p.newInstance(ctx, oauthCtx, instanceURL)
	fullPath, err := i.findRepoFullPathByID(repositoryID)
	if err != nil {
		return "", err
	}

	resp, err := i.APIReadFile(fullPath, filePath, ref)
	if err != nil {
		return "", err
	}

	if resp.Encoding == "base64" {
		content, err := base64.StdEncoding.DecodeString(resp.Content)
		if err != nil {
			return "", errors.Wrap(err, "decode file content")
		}

		pp.Println(string(content))
		return string(content), nil
	}
	return "", errors.Errorf("Unknown encoding %s", resp.Encoding)
}

// GetBranch gets the given branch in the repository.
func (p *Provider) GetBranch(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, repositoryID string, branchName string) (*vcs.BranchInfo, error) {
	i := p.newInstance(ctx, oauthCtx, instanceURL)
	fullPath, err := i.findRepoFullPathByID(repositoryID)
	if err != nil {
		return nil, err
	}

	resp, err := i.APIGetBranch(fullPath, branchName)
	if err != nil {
		return nil, err
	}

	return &vcs.BranchInfo{
		Name:         resp.Name,
		LastCommitID: resp.Commit.Sha,
	}, nil
}

// CreateBranch creates the branch in the repository.
func (p *Provider) CreateBranch(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, repositoryID string, branch *vcs.BranchInfo) error {
	i := p.newInstance(ctx, oauthCtx, instanceURL)
	fullPath, err := i.findRepoFullPathByID(repositoryID)
	if err != nil {
		return err
	}

	if _, err := i.APICreateBranch(fullPath, branch.Name, branch.LastCommitID); err != nil {
		return err
	}

	return nil
}

// ListPullRequestFile lists the changed files in the pull request.
func (p *Provider) ListPullRequestFile(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, repositoryID string, pullRequestID string) ([]*vcs.PullRequestFile, error) {
	i := p.newInstance(ctx, oauthCtx, instanceURL)
	fullPath, err := i.findRepoFullPathByID(repositoryID)
	if err != nil {
		return nil, err
	}

	files, err := i.APIRepoPRFiles(fullPath, pullRequestID)
	if err != nil {
		return nil, err
	}

	ret := make([]*vcs.PullRequestFile, 0)
	for _, f := range files {
		ret = append(ret, &vcs.PullRequestFile{
			Path:         f.Filename,
			LastCommitID: f.Sha,
			IsDeleted:    f.Status == "deleted",
		})
	}

	return ret, nil
}

// CreatePullRequest creates the pull request in the repository.
func (p *Provider) CreatePullRequest(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, repositoryID string, pullRequestCreate *vcs.PullRequestCreate) (*vcs.PullRequest, error) { //nolint:revive
	panic("not implemented") // TODO: Implement
}

// UpsertEnvironmentVariable creates or updates the environment variable in the repository.
func (p *Provider) UpsertEnvironmentVariable(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, repositoryID string, key string, value string) error { //nolint:revive
	// Gitee not provide api to upsert its ci env.
	return nil
}

// CreateWebhook creates a webhook in the repository with given payload.
func (p *Provider) CreateWebhook(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, repositoryID string, payload []byte) (string, error) { //nolint:revive
	panic("not implemented")
}

// PatchWebhook patches the webhook in the repository with given payload.
func (p *Provider) PatchWebhook(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, repositoryID string, webhookID string, payload []byte) error { //nolint:revive
	// TODO: Implement
	return nil
}

// DeleteWebhook deletes the webhook from the repository.
func (p *Provider) DeleteWebhook(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, repositoryID string, webhookID string) error { //nolint:revive
	// TODO: Implement
	return nil
}

// ValidateWebHook validate gitee webhook event.
func (*Provider) ValidateWebHook(secretToken string, httpHeader http.Header, _ []byte) (bool, error) {
	pp.Println(httpHeader.Get("X-Gitee-Event"))

	toHash := fmt.Sprintf("%s\n%s", httpHeader.Get("X-Gitee-Timestamp"), secretToken)

	h := hmac.New(sha256.New, []byte(secretToken))

	if _, err := h.Write([]byte(toHash)); err != nil {
		return false, errors.Wrap(err, "failed to calculate SHA256 of the webhook secret")
	}

	sum := base64.StdEncoding.EncodeToString(h.Sum(nil))

	pp.Println("toHash:", toHash)
	pp.Println("sum:", sum)
	pp.Println("Token:", httpHeader.Get("X-Gitee-Token"))

	return sum == httpHeader.Get("X-Gitee-Token"), nil
}

// ParseWebHook parse gitee webhook event to vcs.PushEvent.
func (*Provider) ParseWebHook(httpHeader http.Header, body []byte) (bool, vcs.PushEvent, error) {
	empty := vcs.PushEvent{}

	pp.Println(httpHeader.Get("X-Gitee-Event"))

	if httpHeader.Get("X-Gitee-Ping") == "true" ||
		httpHeader.Get("X-Gitee-Event") != "Push Hook" {
		return false, empty, nil
	}

	event := &giteePushEvent{}
	if err := json.Unmarshal(body, event); err != nil {
		return false, empty, errors.Wrap(err, "Unmarshal webhook event body fail")
	}
	pp.Println(string(body))
	pp.Println(event)
	pp.Println(event.toVcsPushEvent())

	return true, event.toVcsPushEvent(), nil
}

// CreateWebhookPatch create webhook.
func (p *Provider) CreateWebhookPatch(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, repo vcs.Repository, webHookURL, secretToken string, _ bool) (string, error) {
	// Gitee have no option about ssl.
	resp, err := p.newInstance(ctx, oauthCtx, instanceURL).APICreateWebHook(repo.FullPath, webHookURL, secretToken)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(resp.ID), nil
}

// func (p *Provider) ReadSheetInfo(ctx context.Context, oauthCtx common.OauthContext, instanceURL, repositoryID, filePath, ref string) (*vcs.SheetInfo, error) {
//
//	i := p.newInstance(ctx, oauthCtx, instanceURL)
//
//	fullPath, err := i.findRepoFullPathByID(repositoryID)
//	if err != nil {
//		return nil, err
//	}
//
//	pp.Println(fullPath)
//	file, err := i.APIReadFile(fullPath, filePath, ref)
//	if err != nil {
//		return nil, err
//	}
//	pp.Println(file, err)
//
//	content := ""
//	if file.Encoding == "base64" {
//
//		bts, err := base64.StdEncoding.DecodeString(file.Content)
//		if err != nil {
//			return nil, errors.Wrap(err, "decode file content")
//		}
//		content = string(bts)
//	}
//
//	commit, err := i.APIFileLatestCommit(fullPath, filePath, ref)
//	if err != nil {
//		return nil, err
//	}
//	pp.Println(commit, err)
//
//	return &vcs.SheetInfo{
//		Name:         file.Name,
//		Path:         file.Path,
//		Size:         int64(file.Size),
//		Author:       commit.Author.Name,
//		LastCommitID: commit.Sha,
//		Content:      content,
//	}, nil
//
//}
