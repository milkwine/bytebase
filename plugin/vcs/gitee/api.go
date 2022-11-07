// Package gitee is the plugin for GitHub.
package gitee

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"

	"github.com/k0kubun/pp/v3"
	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/common"
)

// FullPath is needed by Gitee api while `vcs.Interface` only pass repositoryID.
// At this time, find `FullPath` every time when needed.
func (i *instance) findRepoFullPathByID(repositoryID string) (string, error) {
	repos, err := i.APIAllRepoList()
	if err != nil {
		return "", err
	}

	id, err := strconv.Atoi(repositoryID)
	if err != nil {
		return "", err
	}
	for _, repo := range repos {
		if repo.ID == id {
			return repo.FullName, nil
		}
	}

	return "", errors.Errorf("Can't find repo with id %s", repositoryID)
}

func (i *instance) APIAllRepoList() ([]apiRespFetchUserRepo, error) {
	param := struct {
		affiliation string `url:"affiliation"` //nolint:revive
	}{"admin"}

	all := new([]apiRespFetchUserRepo)

	s := i.s().Get("/api/v5/user/repos").QueryStruct(param)
	err := i.pageAll(s, all,
		func() interface{} {
			return new([]apiRespFetchUserRepo)
		},
		func(all interface{}, objs interface{}, page_size int) bool {
			ap := all.(*[]apiRespFetchUserRepo)
			op := objs.(*[]apiRespFetchUserRepo)
			*ap = append(*ap, *op...)

			return len(*op) >= page_size
		})

	if err != nil {
		return nil, err
	}

	return *all, nil
}

func (i *instance) APICreateWebHook(fullName string, webHookURL, secretToken string) (*apiRespCreateWebHook, error) {
	url := fmt.Sprintf("/api/v5/repos/%s/hooks", fullName)
	req := &apiReqCreateWebHook{
		URL:            webHookURL,
		EncryptionType: 1, //0: password, 1: encrypt,
		Password:       secretToken,
		PushEvents:     true,
	}

	resp := &apiRespCreateWebHook{}

	s := i.s().Post(url).BodyJSON(req)

	if err := i.req(s, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (i *instance) APIReadFile(fullName string, filePath, ref string) (*apiRespReadFile, error) {
	url := fmt.Sprintf("/api/v5/repos/%s/contents/%s?ref=%s", fullName, url.QueryEscape(filePath), ref)

	resp := &apiRespReadFile{}

	s := i.s().Get(url)
	pp.Println(s)

	if err := i.req(s, resp); err != nil {
		if match, _ := regexp.MatchString("unrecognized.*code 200", err.Error()); match {
			return nil, common.Errorf(common.NotFound, "failed to read file from URL %s", url)
		}
		return nil, err
	}
	return resp, nil
}

func (i *instance) APIListFile(fullName string, filePath, ref string) ([]apiRespReadFile, error) {
	url := fmt.Sprintf("/api/v5/repos/%s/contents/%s?ref=%s", fullName, url.QueryEscape(filePath), ref)

	resp := new([]apiRespReadFile)

	s := i.s().Get(url)

	if err := i.req(s, resp); err != nil {
		return nil, err
	}
	return *resp, nil
}

func (i *instance) APILoginUser() (*apiRespUser, error) {
	resp := &apiRespUser{}

	s := i.s().Get("/api/v5/user")

	if err := i.req(s, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (i *instance) APILoginEmail() ([]apiRespEmail, error) {
	resp := new([]apiRespEmail)

	s := i.s().Get("/api/v5/emails")

	if err := i.req(s, resp); err != nil {
		return nil, err
	}
	return *resp, nil
}

func (i *instance) APIFileLatestCommit(fullName string, filePath, ref string) (*apiRespRepoCommit, error) {
	resp := new([]apiRespRepoCommit)
	url := fmt.Sprintf("/api/v5/repos/%s/commits", fullName)

	param := struct {
		path     string `url:"path"`     //nolint:revive
		ref      string `url:"sha"`      //nolint:revive
		page     int    `url:"page"`     //nolint:revive
		per_page int    `url:"per_page"` //nolint:revive
	}{filePath, ref, 1, 1}

	s := i.s().Get(url).QueryStruct(&param)
	pp.Println(s)

	if err := i.req(s, resp); err != nil {
		return nil, err
	}
	if len(*resp) == 0 {
		return nil, errors.Errorf("Can't find commit info with file %s in ref %s", filePath, ref)
	}
	return &(*resp)[0], nil
}

func (i *instance) APIRepoMember(fullName string) ([]apiRespRepoMember, error) {
	all := new([]apiRespRepoMember)

	url := fmt.Sprintf("/api/v5/repos/%s/collaborators", fullName)
	s := i.s().Get(url)

	err := i.pageAll(s, all,
		func() interface{} { return new([]apiRespRepoMember) }, //nolint:revive
		func(all interface{}, objs interface{}, page_size int) bool {
			ap := all.(*[]apiRespRepoMember)
			op := objs.(*[]apiRespRepoMember)
			*ap = append(*ap, *op...)

			return len(*op) >= page_size
		})
	if err != nil {
		return nil, err
	}
	return *all, nil
}

func (i *instance) APIRepoPRFiles(fullName string, pullRequestID string) ([]apiRespPRFile, error) {
	url := fmt.Sprintf("/api/v5/repos/%s/pulls/%s/files", fullName, pullRequestID)
	s := i.s().Get(url)

	resp := new([]apiRespPRFile)
	if err := i.req(s, resp); err != nil {
		return nil, err
	}
	return *resp, nil
}

func (i *instance) APICreateBranch(fullName string, branchName, fromRef string) (*apiRespBranch, error) {
	req := struct {
		Refs       string `json:"refs"`
		BranchName string `json:"branch_name"`
	}{fromRef, branchName}

	resp := &apiRespBranch{}

	s := i.s().Post(fmt.Sprintf("/api/v5/repos/%s/branches", fullName)).BodyJSON(&req)

	if err := i.req(s, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (i *instance) APIGetBranch(fullName string, branchName string) (*apiRespBranch, error) {
	resp := &apiRespBranch{}

	s := i.s().Get(fmt.Sprintf("/api/v5/repos/%s/branches/%s", fullName, branchName))

	if err := i.req(s, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (i *instance) APIGetCommit(fullName string, commitID string) (*apiRespRepoCommit, error) {
	resp := &apiRespRepoCommit{}

	s := i.s().Get(fmt.Sprintf("/api/v5/repos/%s/commits/%s", fullName, commitID))

	if err := i.req(s, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
