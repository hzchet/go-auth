package controller

import (
	"context"
	"net/http"

	"server/internal/pkg/tokens"
	"server/internal/utils"
	"server/internal/pkg/metrics"

	"github.com/juju/zaputil/zapctx"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Controller struct {
	Config utils.ConfigScheme
}

func New(configPath string, ctx context.Context) *Controller {
	logger := zapctx.Logger(ctx)
	
	_, span := metrics.Tracer.Start(ctx, "New")
	defer span.End()

	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		logger.Error("can't read config", zap.Error(err))
	}

	controller := Controller{}
	if err := viper.Unmarshal(&controller.Config); err != nil {
		logger.Error("unmarshaling of the config file failed", zap.Error(err))
	}

	return &controller
}

func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	logger := zapctx.Logger(r.Context())
	logger.Info("Login handler")

	newCtx, span := metrics.Tracer.Start(r.Context(), "Login")
	defer span.End()

	username, password, _ := r.BasicAuth()

	if value, ok := c.Config.Users[username]; ok && value.IsEqual(password) {
		accessToken, refreshToken := tokens.NewPair(newCtx, username)

		tokens.AddToCookies(newCtx, w, accessToken, "access_token")
		tokens.AddToCookies(newCtx, w, refreshToken, "refresh_token")

		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

func (c *Controller) Verify(w http.ResponseWriter, r *http.Request) {
	logger := zapctx.Logger(r.Context())
	
	newCtx, span := metrics.Tracer.Start(r.Context(), "Verify")
	defer span.End()

	accessToken, issuer, err := tokens.ExtractFromCookies(r, "access_token")

	if err == nil && accessToken.Valid {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(issuer.(string)))
	} else {
		logger.Debug("accessToken invalid", zap.Error(err))
		refreshToken, issuer, err := tokens.ExtractFromCookies(r, "refresh_token")

		if err == nil && refreshToken.Valid {
			username := issuer.(string)

			accessToken, refreshToken := tokens.NewPair(newCtx, username)

			tokens.AddToCookies(newCtx, w, accessToken, "access_token")
			tokens.AddToCookies(newCtx, w, refreshToken, "refresh_token")

			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	}
}
