package tokens

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	privateKey = []byte(os.Getenv("JWT_PRIVATE_KEY"))
)

func NewPair(issuer string) (string, string) {
	return generate(1, issuer), generate(60, issuer)
}

func AddToCookies(w http.ResponseWriter, token string, name string) {
	cookie := &http.Cookie{
		Name:  name,
		Value: token,
	}
	http.SetCookie(w, cookie)
}

func ExtractFromCookies(r *http.Request, name string) (*jwt.Token, interface{}, error) {
	tokenCookie, err := r.Cookie(name)
	if err != nil {
		log.Fatalf("error occurred while reading cookie: %v\n", err)
	}
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		tokenCookie.Value,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return privateKey, nil
		},
	)
	return token, claims["Issuer"], err
}

func IsExpired(err error) bool {
	return errors.Is(err, jwt.ErrTokenExpired)
}

func generate(duration time.Duration, issuer string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"ExpiresAt": jwt.NewNumericDate(time.Now().Add(duration * time.Minute)),
		"IssuedAt":  jwt.NewNumericDate(time.Now()),
		"Issuer":    issuer,
	})
	signedToken, _ := token.SignedString(privateKey)
	return signedToken
}
