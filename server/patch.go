package server

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/api"
	"github.com/bytebase/bytebase/common"
	vcsPlugin "github.com/bytebase/bytebase/plugin/vcs"
)

func (s *Server) createVCSWebhookPatch(ctx context.Context, vcsType vcsPlugin.Type, instanceURL string, repoCreate api.RepositoryCreate) (string, error) {
	repoExID, err := strconv.Atoi(repoCreate.ExternalID)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to parse repository id %s", repoCreate.ExternalID)
	}

	vcsRepo := vcsPlugin.Repository{
		ID:       int64(repoExID),
		Name:     repoCreate.Name,
		FullPath: repoCreate.FullPath,
		WebURL:   repoCreate.WebURL,
	}

	url := fmt.Sprintf("%s/hook/%s/%s", s.profile.ExternalURL, vcsType.RouterPrefix(), repoCreate.WebhookEndpointID)

	oauthCtx := common.OauthContext{
		AccessToken: repoCreate.AccessToken,
		// We use refreshTokenNoop() because the repository isn't created yet.
		Refresher: refreshTokenNoop(),
	}

	webhookID, err := vcsPlugin.Get(vcsType, vcsPlugin.ProviderConfig{}).CreateWebhookPatch(
		ctx,
		oauthCtx,
		instanceURL,
		vcsRepo,
		url,
		repoCreate.WebhookSecretToken,
		false,
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to create webhook")
	}
	return webhookID, nil
}
