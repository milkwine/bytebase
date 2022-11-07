package gitee

import (
	"context"
	"net/http"
	"testing"

	"github.com/k0kubun/pp/v3"

	"github.com/bytebase/bytebase/common"
)

func TestFunction(_ *testing.T) {
	p := &Provider{http.DefaultClient}

	ctx := context.Background()
	oauthCtx := common.OauthContext{
		AccessToken:  "",
		RefreshToken: "",
	}
	instanceURL := "https://gitee.com"

	i := p.newInstance(ctx, oauthCtx, instanceURL)
	resp, err := i.APIAllRepoList()
	pp.Println(err, resp)
}

func TestValidateWebHook(_ *testing.T) {
	p := &Provider{http.DefaultClient}

	head := http.Header{}

	head.Add("X-Gitee-Token", "4qolhWRjWWbiSNB59b/l11gMjxt9Wv4F1TNnkUL2oOs=")
	head.Add("X-Gitee-Timestamp", "1667483756719")

	isValid, err := p.ValidateWebHook("EkP6kswO9m7cvBw0", head, []byte{})
	pp.Println(err, isValid)
}
