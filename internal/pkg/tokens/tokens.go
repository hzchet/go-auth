package tokens

import (
	"net/http"
	"context"
	"os"
	"time"

	"server/internal/pkg/metrics"

	"github.com/golang-jwt/jwt/v4"
	"github.com/juju/zaputil/zapctx"
	"go.uber.org/zap"
)

var (
	privateKey = []byte(os.Getenv("JWT_PRIVATE_KEY"))
)

func NewPair(ctx context.Context, issuer string) (string, string) {
	newCtx, span := metrics.Tracer.Start(ctx, "AddToCookies")
	defer span.End()
	return generate(newCtx, 1, issuer), generate(newCtx, 60, issuer)
}

func AddToCookies(ctx context.Context, w http.ResponseWriter, token string, name string) {
	_, span := metrics.Tracer.Start(ctx, "AddToCookies")
	defer span.End()

	cookie := &http.Cookie{
		Name:  name,
		Value: token,
	}

	http.SetCookie(w, cookie)
}

func ExtractFromCookies(r *http.Request, name string) (*jwt.Token, interface{}, error) {
	logger := zapctx.Logger(r.Context())
	_, span := metrics.Tracer.Start(r.Context(), "ExtractFromCookies")
	defer span.End()

	tokenCookie, err := r.Cookie(name)

	if err != nil {
		logger.Debug("error occurred while reading cookie", zap.Error(err))
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

func generate(ctx context.Context, duration time.Duration, issuer string) string {
	logger := zapctx.Logger(ctx)
	_, span := metrics.Tracer.Start(ctx, "generate")
	defer span.End()
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"ExpiresAt": jwt.NewNumericDate(time.Now().Add(duration * time.Minute)),
		"IssuedAt":  jwt.NewNumericDate(time.Now()),
		"Issuer":    issuer,
	})
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		logger.Debug("error occurred while signing token", zap.Error(err))
	}
	return signedToken
}
