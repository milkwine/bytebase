package vcs

import "strings"

// RouterPrefix return router path prefix according to vcs type.
func (t Type) RouterPrefix() string {
	switch t {
	case GitLabSelfHost:
		return "gitlab"
	case GitHubCom:
		return "github"
	default:
		return strings.ReplaceAll(strings.ToLower(string(t)), "_", "-")
	}
}
