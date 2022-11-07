package server

import (
	"server/internal/pkg/controller"
	"server/internal/server"

	"errors"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"
	"encoding/base64"

	"github.com/stretchr/testify/suite"
)

type integraTestSuite struct {
	suite.Suite
	app server.HttpServer
	authClient *http.Client
}

func TestIntegraTestSuite(t *testing.T) {
	suite.Run(t, &integraTestSuite{})
}

func (s *integraTestSuite) SetupSuite() {
	s.app = server.HttpServer{}
	s.authClient = &http.Client{Timeout: 10 * time.Second}
	go func() {
		if err := s.app.Start("../../config"); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("start server failed: %#v", err)
		}
	}()
}

func (s *integraTestSuite) TearDownSuite() {
	err := s.app.Stop(context.Background())
	if err != nil {
		log.Fatalf("server shutdown failed: %#v", err)
	}
}

func (s *integraTestSuite) TestLoginCorrect() {
	loginEndpoint := "http://localhost:8080/auth/api/v1/login"
	req, err := http.NewRequest(http.MethodPost, loginEndpoint, nil)

	if err != nil {
		log.Fatal("login failed")
	}

	email := "email1"
	password := "password1"
	str := fmt.Sprintf("%s:%s", email, password)
	str = base64.StdEncoding.EncodeToString([]byte(str))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", str))

	resp, err := s.authClient.Do(req)
	s.NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)
}

func (s *integraTestSuite) TestLoginInCorrectPassword() {
	loginEndpoint := "http://localhost:8080/auth/api/v1/login"
	req, err := http.NewRequest(http.MethodPost, loginEndpoint, nil)

	if err != nil {
		log.Fatal("login failed")
	}

	email := "email1"
	password := "incorrectPassword"
	str := fmt.Sprintf("%s:%s", email, password)
	str = base64.StdEncoding.EncodeToString([]byte(str))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", str))

	resp, err := s.authClient.Do(req)
	s.NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusForbidden, resp.StatusCode)
}

func (s *integraTestSuite) TestLoginInCorrectEmail() {
	loginEndpoint := "http://localhost:8080/auth/api/v1/login"
	req, err := http.NewRequest(http.MethodPost, loginEndpoint, nil)

	if err != nil {
		log.Fatal("login failed")
	}

	email := "incorrectEmail"
	password := "password1"
	str := fmt.Sprintf("%s:%s", email, password)
	str = base64.StdEncoding.EncodeToString([]byte(str))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", str))

	resp, err := s.authClient.Do(req)
	s.NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusForbidden, resp.StatusCode)
}

func (s *integraTestSuite) TestVerifyCorrect() {
	loginEndpoint := "http://localhost:8080/auth/api/v1/login"
	req, err := http.NewRequest(http.MethodPost, loginEndpoint, nil)

	if err != nil {
		log.Fatal("login failed")
	}

	email := "email2"
	password := "password2"
	str := fmt.Sprintf("%s:%s", email, password)
	str = base64.StdEncoding.EncodeToString([]byte(str))

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", str))
	respLogin, err := s.authClient.Do(req)
	s.NoError(err)
	defer respLogin.Body.Close()
	s.Equal(http.StatusOK, respLogin.StatusCode)
	
	cookies := respLogin.Cookies()
	s.GreaterOrEqual(len(cookies), 2)

	accessCookie := cookies[len(cookies) - 2]
	refreshCookie := cookies[len(cookies) - 1]

	verifyEndpoint := "http://localhost:8080/auth/api/v1/verify"
	req, err = http.NewRequest(http.MethodPost, verifyEndpoint, nil)
	s.NoError(err)
	req.AddCookie(accessCookie)
	req.AddCookie(refreshCookie)
	respVerify, err := s.authClient.Do(req)
	s.NoError(err)
	defer respVerify.Body.Close()
	s.Equal(http.StatusOK, respVerify.StatusCode)

	user := controller.UserProjection{}
	json.NewDecoder(respVerify.Body).Decode(&user)
	s.Equal("email2", user.Email)
}

func (s *integraTestSuite) TestVerifyIncorrectAccessToken() {
	loginEndpoint := "http://localhost:8080/auth/api/v1/login"
	req, err := http.NewRequest(http.MethodPost, loginEndpoint, nil)

	if err != nil {
		log.Fatal("login failed")
	}

	email := "email3"
	password := "password3"
	str := fmt.Sprintf("%s:%s", email, password)
	str = base64.StdEncoding.EncodeToString([]byte(str))

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", str))
	respLogin, err := s.authClient.Do(req)
	s.NoError(err)
	defer respLogin.Body.Close()
	s.Equal(http.StatusOK, respLogin.StatusCode)
	
	cookies := respLogin.Cookies()
	s.GreaterOrEqual(len(cookies), 2)

	accessCookie := &http.Cookie{
		Name: "access_token",
		Value: "wrong access token!!",
	}
	refreshCookie := cookies[len(cookies) - 1]  // correct refreshToken

	verifyEndpoint := "http://localhost:8080/auth/api/v1/verify"
	req, err = http.NewRequest(http.MethodPost, verifyEndpoint, nil)
	s.NoError(err)
	req.AddCookie(accessCookie)
	req.AddCookie(refreshCookie)
	respVerify, err := s.authClient.Do(req)
	s.NoError(err)
	defer respVerify.Body.Close()
	s.Equal(http.StatusOK, respVerify.StatusCode)

	user := controller.UserProjection{}
	json.NewDecoder(respVerify.Body).Decode(&user)
	s.Equal("email3", user.Email)
}

func (s *integraTestSuite) TestVerifyIncorrectTokens() {
	loginEndpoint := "http://localhost:8080/auth/api/v1/login"
	req, err := http.NewRequest(http.MethodPost, loginEndpoint, nil)

	if err != nil {
		log.Fatal("login failed")
	}

	email := "email3"
	password := "password3"
	str := fmt.Sprintf("%s:%s", email, password)
	str = base64.StdEncoding.EncodeToString([]byte(str))

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", str))
	respLogin, err := s.authClient.Do(req)
	s.NoError(err)
	defer respLogin.Body.Close()
	s.Equal(http.StatusOK, respLogin.StatusCode)
	
	accessCookie := &http.Cookie{
		Name: "access_token",
		Value: "wrong access token!!",
	}
	refreshCookie := &http.Cookie{
		Name: "refresh_token",
		Value: "wrong refresh token!",
	}

	verifyEndpoint := "http://localhost:8080/auth/api/v1/verify"
	req, err = http.NewRequest(http.MethodPost, verifyEndpoint, nil)
	s.NoError(err)
	req.AddCookie(accessCookie)
	req.AddCookie(refreshCookie)
	respVerify, err := s.authClient.Do(req)
	s.NoError(err)
	defer respVerify.Body.Close()
	s.Equal(http.StatusForbidden, respVerify.StatusCode)
}
