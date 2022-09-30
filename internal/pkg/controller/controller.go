package controller

import (
	"log"
	"net/http"
	
	"server/internal/pkg/tokens"
	"server/internal/utils"

	"github.com/spf13/viper"
)

type Controller struct {
	Config *utils.ConfigScheme
}

func New(configPath string) *Controller {
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("fatal error occurred while reading config file: %v", err)
	}

	controller := Controller{}
	if err := viper.Unmarshal(controller.Config); err != nil {
		log.Fatalf("unable to decode config file into struct: %v", err)
	}

	return &controller
}

func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	username, password, _ := r.BasicAuth()

	if value, ok := c.Config.Users[username]; ok && value.IsEqual(password) {
		accessToken, refreshToken := tokens.NewPair(username)

		tokens.AddToCookies(w, accessToken, "access_token")
		tokens.AddToCookies(w, refreshToken, "refresh_token")

		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

func (c *Controller) Verify(w http.ResponseWriter, r *http.Request) {
	accessToken, issuer, err := tokens.ExtractFromCookies(r, "access_token")

	switch {
	case accessToken.Valid:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(issuer.(string)))

	case tokens.IsExpired(err):
		refreshToken, issuer, _ := tokens.ExtractFromCookies(r, "refresh_token")

		if refreshToken.Valid {
			username := issuer.(string)

			accessToken, refreshToken := tokens.NewPair(username)

			tokens.AddToCookies(w, accessToken, "access_token")
			tokens.AddToCookies(w, refreshToken, "refresh_token")

			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	}
}
