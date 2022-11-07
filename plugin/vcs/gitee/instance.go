package gitee

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dghubble/sling"
	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/common"
)

type instance struct {
	ctx         context.Context
	doer        *http.Client
	oauthCtx    common.OauthContext
	instanceURL string
	maxRetries  int
	pageSize    int
}

func (p *Provider) newInstance(ctx context.Context, oauthCtx common.OauthContext, instanceURL string) *instance {
	return &instance{ctx: ctx, doer: p.client, oauthCtx: oauthCtx, instanceURL: instanceURL,
		maxRetries: maxRetries, pageSize: pageSize}
}

func (i *instance) s() *sling.Sling {
	return sling.New().Doer(i.doer).Base(i.instanceURL)
}

// Wrap retry and token refresh.
func (i *instance) req(s *sling.Sling, sucObj interface{}) error {
	var lastE error
	var lastCode int

	for retries := 0; retries < i.maxRetries; retries++ {
		select {
		case <-i.ctx.Done():
			return i.ctx.Err()
		default:
		}

		request, err := s.New().Add("Authorization", fmt.Sprintf("Bearer %s", i.oauthCtx.AccessToken)).Request()
		if err != nil {
			return errors.Wrap(err, "setup http request fail")
		}

		request = request.WithContext(i.ctx)

		resp, err := i.doer.Do(request)
		if err != nil {
			lastCode, lastE = 0, errors.Wrap(err, "http request fail")
			continue
		}

		bts, err := io.ReadAll(resp.Body)
		if err != nil {
			lastCode, lastE = resp.StatusCode, errors.Wrap(err, "http read body fail")
			continue
		}
		_ = resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			if err := json.Unmarshal(bts, sucObj); err != nil {
				return errors.Errorf("unrecognized response with code %d body %s", resp.StatusCode, string(bts))
			}
			return nil
		}

		oe := &oauthError{}
		if err := json.Unmarshal(bts, oe); err == nil && oe.Err != "" {
			lastCode, lastE = resp.StatusCode, oe

			// token refresh
			if (oe.Err == "invalid_token" || oe.Err == "invalid_grant") && strings.Contains(oe.ErrorDescription, "expired") {
				if err := i.freshToken(); err != nil {
					lastCode, lastE = 0, err
				}
			}
			continue
		}

		ae := &apiError{}
		if err := json.Unmarshal(bts, ae); err == nil && (ae.Message != "" || len(ae.Messages) > 0) {
			lastCode, lastE = resp.StatusCode, ae

			// token refresh
			if strings.Contains(ae.Message, "expired") {
				if err := i.freshToken(); err != nil {
					lastCode, lastE = 0, err
				}
			}
			continue
		}

		return errors.Errorf("unrecognized response with code %d body %s", resp.StatusCode, string(bts))
	}

	if lastCode != 0 {
		return errors.Wrapf(lastE, "retries exceeded with status code %d", lastCode)
	}
	return errors.Wrap(lastE, "retries exceeded")
}

func (i *instance) freshToken() error {
	params := fmt.Sprintf("grant_type=refresh_token&refresh_token=%s", i.oauthCtx.RefreshToken)

	sucObj, err := oauthGrant(i.ctx, i.doer, i.instanceURL, params)
	if err != nil {
		return err
	}

	if sucObj == nil {
		return errors.Errorf("refresh OAuth token fail")
	}

	i.oauthCtx.AccessToken = sucObj.AccessToken
	i.oauthCtx.RefreshToken = sucObj.RefreshToken
	if err := i.oauthCtx.Refresher(sucObj.AccessToken, sucObj.RefreshToken, sucObj.CreatedAt+sucObj.ExpiresIn); err != nil {
		return errors.Wrap(err, "call fresher fail")
	}
	return nil
}

type pageNewSetFunc func() interface{}
type pageAppendFunc func(all interface{}, objs interface{}, page_size int) bool

func (i *instance) pageAll(s *sling.Sling, all interface{}, newSetFunc pageNewSetFunc, appendFunc pageAppendFunc) error {
	param := struct {
		page     int `url:"page"`     //nolint:revive
		per_page int `url:"per_page"` //nolint:revive
	}{1, i.pageSize}

	for {
		objs := newSetFunc()
		ns := s.New().QueryStruct(param)
		if err := i.req(ns, objs); err != nil {
			return err
		}

		if !appendFunc(all, objs, param.per_page) {
			break
		}
		param.page++
	}
	return nil
}

type apiError struct {
	Messages []string `json:"messages"`
	Message  string   `json:"message"`
}

func (e apiError) Error() string {
	msg := make([]string, 0)
	if len(e.Messages) > 0 {
		msg = append(msg, e.Messages...)
	}
	if e.Message != "" {
		msg = append(msg, e.Message)
	}
	return fmt.Sprintf("Api response error message [%s]", strings.Join(msg, ","))
}
