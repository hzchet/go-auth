package main

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
	"golang.org/x/crypto/pbkdf2"
)

type httpServer struct {
	server *http.Server
}

type Password string

func (p Password) IsEqual(otherPassword string) bool {
	hashed := pbkdf2.Key([]byte(otherPassword), []byte(os.Getenv("SALT")), 4096, 32, sha1.New)
	return p == Password(base64.StdEncoding.EncodeToString(hashed))
}

type Config struct {
	Port int64
	Users map[string]Password
}

func generateToken(duration time.Duration, issuer string, privateKey []byte) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"ExpiresAt": jwt.NewNumericDate(time.Now().Add(duration * time.Minute)),
		"IssuedAt": jwt.NewNumericDate(time.Now()),
		"Issuer": issuer,
	})
	signedToken, _ := token.SignedString(privateKey)
	return signedToken
}

func addTokenToCookies(w http.ResponseWriter, token string, name string) {
	cookie := &http.Cookie{
		Name: name,
		Value: token,
	}
	http.SetCookie(w, cookie)
}

func login(config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, _ := r.BasicAuth()
		
		if value, ok := config.Users[username]; ok && value.IsEqual(password) {
			privateKey := []byte(os.Getenv("JWT_PRIVATE_KEY"))

			accessToken := generateToken(1, username, privateKey)
			refreshToken := generateToken(60, username, privateKey)
			
			addTokenToCookies(w, accessToken, "access_token")
			addTokenToCookies(w, refreshToken, "refresh_token")
			
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	}
}

func extractTokenFromCookies(r *http.Request, name string, privateKey []byte) (jwt.MapClaims, 
	*jwt.Token, error) {
	tokenCookie, err := r.Cookie(name)
	if err != nil {
		log.Fatalf("error occurred while reading cookie: %v\n", err)
	}
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		tokenCookie.Value, 
		claims, 
		func (token *jwt.Token) (interface{}, error) {
			return privateKey, nil
		},
	)
	return claims, token, err
}

func verify(w http.ResponseWriter, r *http.Request) {
	privateKey := []byte(os.Getenv("JWT_PRIVATE_KEY"))
	accessClaims, accessToken, err := extractTokenFromCookies(r, "access_token", privateKey)
	if accessToken.Valid {
		w.WriteHeader(http.StatusOK)
		if issuer, ok := accessClaims["Issuer"]; ok {
			w.Write([]byte(issuer.(string)))
		}
	} else if errors.Is(err, jwt.ErrTokenExpired) {
		refreshClaims, refreshToken, _ := extractTokenFromCookies(r, "refresh_token", privateKey)
		if refreshToken.Valid {
			username := refreshClaims["Issuer"].(string)
			
			accessToken := generateToken(1, username, privateKey)
			refreshToken := generateToken(60, username, privateKey)

			addTokenToCookies(w, accessToken, "access_token")
			addTokenToCookies(w, refreshToken, "refresh_token")
			
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	}
}

func (s *httpServer) Start() error {
	viper.SetConfigFile("config/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("fatal error occurred while reading config file: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("unable to decode config file into struct: %v", err)
	}
	
	router := chi.NewRouter()
	router.Post("/auth/api/v1/login", login(&config))
	router.Post("/auth/api/v1/verify", verify)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: router,
	}

	return s.server.ListenAndServe()
}

func (s *httpServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	server := httpServer{}
	go func() {
		if err := server.Start(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("start server failed: %#v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 1 * time.Minute)
	defer cancel()

	err := server.Stop(shutdownCtx)
	if err != nil {
		log.Fatalf("server shutdown failed: %#v", err)
	}
}
