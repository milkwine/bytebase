package github

import (
	"context"
	"net/http"

	"github.com/bytebase/bytebase/common"
	"github.com/bytebase/bytebase/plugin/vcs"
)

func (*Provider) ValidateWebHook(secretToken string, httpHeader http.Header, body []byte) (bool, error) { //nolint:revive
	panic("not implemente yet")
}
func (*Provider) ParseWebHook(httpHead http.Header, body []byte) (bool, vcs.PushEvent, error) { //nolint:revive
	panic("not implemente yet")
}

func (*Provider) CreateWebhookPatch(ctx context.Context, oauthCtx common.OauthContext, instanceURL string, repo vcs.Repository, webHookURL, secretToken string, secureSSL bool) (string, error) { //nolint:revive
	panic("not implemente yet")
}
