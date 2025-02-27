package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/jsonapi"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bytebase/bytebase/api"
	"github.com/bytebase/bytebase/common"
	"github.com/bytebase/bytebase/common/log"
	enterpriseAPI "github.com/bytebase/bytebase/enterprise/api"
)

func (s *Server) registerSubscriptionRoutes(g *echo.Group) {
	g.GET("/subscription", func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, &s.subscription); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to marshal subscription response").SetInternal(err)
		}
		return nil
	})

	g.PATCH("/subscription", func(c echo.Context) error {
		patch := &enterpriseAPI.SubscriptionPatch{}
		if err := jsonapi.UnmarshalPayload(c.Request().Body, patch); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Malformed create subscription request").SetInternal(err)
		}
		patch.UpdaterID = c.Get(getPrincipalIDContextKey()).(int)

		if err := s.LicenseService.StoreLicense(patch); err != nil {
			if common.ErrorCode(err) == common.Invalid {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to store license").SetInternal(err)
		}

		ctx := c.Request().Context()
		s.subscription = s.loadSubscription(ctx)

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, &s.subscription); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to marshal subscription response").SetInternal(err)
		}
		return nil
	})

	g.POST("/subscription/trial", func(c echo.Context) error {
		create := &api.TrialPlanCreate{}
		if err := jsonapi.UnmarshalPayload(c.Request().Body, create); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Malformed create trial request").SetInternal(err)
		}

		license := &enterpriseAPI.License{
			InstanceCount: create.InstanceCount,
			ExpiresTs:     time.Now().AddDate(0, 0, create.Days).Unix(),
			IssuedTs:      time.Now().Unix(),
			Plan:          create.Type,
			// the subject format for license should be {org id in hub}.{subscription id in hub}
			// as we just need to simply generate the trialing license in console, we can use the workspace id instead.
			Subject:  fmt.Sprintf("%s.%s", s.workspaceID, ""),
			Trialing: true,
			OrgName:  s.workspaceID,
		}

		value, err := json.Marshal(license)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to marshal license").SetInternal(err)
		}

		ctx := c.Request().Context()
		if _, err := s.store.CreateSettingIfNotExist(ctx, &api.SettingCreate{
			CreatorID:   c.Get(getPrincipalIDContextKey()).(int),
			Name:        api.SettingEnterpriseTrial,
			Value:       string(value),
			Description: "The trialing license.",
		}); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create license").SetInternal(err)
		}

		s.subscription = s.loadSubscription(ctx)

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		if err := jsonapi.MarshalPayload(c.Response().Writer, &s.subscription); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to marshal subscription response").SetInternal(err)
		}
		return nil
	})
}

// loadLicense will load current subscription by license.
// Return subscription with free plan if no license found.
func (s *Server) loadSubscription(ctx context.Context) enterpriseAPI.Subscription {
	subscription := enterpriseAPI.Subscription{
		Plan: api.FREE,
		// -1 means not expire, just for free plan
		ExpiresTs:     -1,
		InstanceCount: 5,
	}

	license, _ := s.loadLicense(ctx)
	if license != nil {
		subscription = enterpriseAPI.Subscription{
			Plan:          license.Plan,
			ExpiresTs:     license.ExpiresTs,
			StartedTs:     license.IssuedTs,
			InstanceCount: license.InstanceCount,
			Trialing:      license.Trialing,
			OrgID:         license.OrgID(),
			OrgName:       license.OrgName,
		}
	}

	return subscription
}

// loadLicense will get and parse valid license from file.
func (s *Server) loadLicense(ctx context.Context) (*enterpriseAPI.License, error) {
	license, err := s.LicenseService.LoadLicense()
	if license != nil {
		log.Info(
			"Load valid license",
			zap.String("plan", license.Plan.String()),
			zap.Time("expiresAt", time.Unix(license.ExpiresTs, 0)),
			zap.Int("instanceCount", license.InstanceCount),
		)

		return license, nil
	}

	if common.ErrorCode(err) == common.NotFound {
		log.Debug("Failed to find license", zap.String("error", err.Error()))
	} else {
		log.Warn("Failed to load valid license", zap.String("error", err.Error()))
	}

	// find free trial license in console
	settingName := api.SettingEnterpriseTrial
	settings, err := s.store.FindSetting(ctx, &api.SettingFind{
		Name: &settingName,
	})
	if err != nil {
		log.Warn("Failed to load trial license from settings", zap.String("error", err.Error()))
		return nil, err
	}
	if len(settings) == 0 {
		return nil, common.Wrapf(err, common.NotFound, "cannot find license")
	}

	var data enterpriseAPI.License
	if err := json.Unmarshal([]byte(settings[0].Value), &data); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal value %q", settings[0].Value)
	}

	return &data, nil
}

func (s *Server) feature(feature api.FeatureType) bool {
	return api.Feature(feature, s.getEffectivePlan())
}

func (s *Server) getPlanLimitValue(name api.PlanLimit) int64 {
	v, ok := api.PlanLimitValues[name]
	if !ok {
		return 0
	}
	return v[s.getEffectivePlan()]
}

func (s *Server) getEffectivePlan() api.PlanType {
	if expireTime := time.Unix(s.subscription.ExpiresTs, 0); expireTime.Before(time.Now()) {
		return api.FREE
	}
	return s.subscription.Plan
}
