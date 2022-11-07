package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"server/internal/pkg/metrics"
	"server/internal/pkg/tokens"
	"server/internal/utils"

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

	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
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

type UserProjection struct {
	Email string
}

type ErrorProjection struct {
	Error string
}

func (c *Controller) Verify(w http.ResponseWriter, r *http.Request) {
	logger := zapctx.Logger(r.Context())
	
	newCtx, span := metrics.Tracer.Start(r.Context(), "Verify")
	defer span.End()

	accessToken, issuer, err := tokens.ExtractFromCookies(r, "access_token")

	if err == nil && accessToken.Valid {
		user := UserProjection{Email: issuer.(string)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	} else {
		logger.Debug("accessToken invalid", zap.Error(err))
		refreshToken, issuer, err := tokens.ExtractFromCookies(r, "refresh_token")

		if err == nil && refreshToken.Valid {
			username := issuer.(string)

			accessToken, refreshToken := tokens.NewPair(newCtx, username)

			tokens.AddToCookies(newCtx, w, accessToken, "access_token")
			tokens.AddToCookies(newCtx, w, refreshToken, "refresh_token")

			user := UserProjection{Email: issuer.(string)}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(user)
		} else {
			err := ErrorProjection{Error: "access denied"}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(err)
		}
	}
}
