package gitee

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
	"github.com/k0kubun/pp/v3"
	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/plugin/vcs"
)

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	CreatedAt    int64  `json:"created_at"`
}

func (t tokenResponse) toVCSOAuthToken() *vcs.OAuthToken {
	oauthToken := &vcs.OAuthToken{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		ExpiresIn:    t.ExpiresIn,
		CreatedAt:    t.CreatedAt,
	}

	if oauthToken.ExpiresIn != 0 {
		oauthToken.ExpiresTs = oauthToken.CreatedAt + oauthToken.ExpiresIn
	}
	return oauthToken
}

type oauthError struct {
	Err              string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e oauthError) Error() string {
	return fmt.Sprintf("OAuth response error %q description %q", e.Err, e.ErrorDescription)
}

func oauthGrant(ctx context.Context, doer *http.Client, instanceURL, params string) (*tokenResponse, error) {
	// Get Access token: /oauth/token?grant_type=authorization_code&code={code}&client_id={client_id}&redirect_uri={redirect_uri}&client_secret={client_secret}
	// Fresh Access token: /oauth/token?grant_type=refresh_token&refresh_token={fresh_token}

	s := sling.New().Doer(doer).Base(instanceURL).Post(fmt.Sprintf("/oauth/token?%s", params))

	pp.Println(s)

	req, err := s.Request()
	if err != nil {
		return nil, errors.Wrap(err, "failed to grant OAuth")
	}

	req = req.WithContext(ctx)

	code, sucObj, oe := 0, &tokenResponse{}, &oauthError{}
	resp, err := s.Do(req, sucObj, oe)

	pp.Println(sucObj, oe)

	if resp != nil {
		code = resp.StatusCode
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to grant OAuth with code %d", code)
	}

	if oe.Err != "" {
		return nil, errors.Wrapf(oe, "failed to grant OAuth token with code %d", code)
	}

	if sucObj.AccessToken != "" {
		return sucObj, nil
	}

	return nil, errors.Errorf("failed to grant OAuth token with code %d (unrecognized response)", code)
}
