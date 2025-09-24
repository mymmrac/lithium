package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/mymmrac/lithium/pkg/module/id"
	"github.com/mymmrac/lithium/pkg/module/user"
)

const tokenCookie = "token"

var userKey = userKeyType{} //nolint:gochecknoglobals

type userKeyType struct{}

type User struct {
	ID id.ID
}

func UserFromContext(ctx context.Context) (User, bool) {
	authUser, ok := ctx.Value(userKey).(User)
	if !ok || authUser.ID == 0 {
		return User{}, false
	}
	return authUser, true
}

type Auth interface {
	Middleware(fCtx fiber.Ctx) error
	GenerateAndSetToken(fCtx fiber.Ctx, userModel *user.Model) error
	ClearToken(fCtx fiber.Ctx)
}

type auth struct {
	cfg Config
}

func NewAuth(cfg Config) Auth {
	return &auth{
		cfg: cfg,
	}
}

func (a *auth) Middleware(fCtx fiber.Ctx) error {
	token := fCtx.Cookies(tokenCookie)
	if token == "" {
		return fiber.ErrUnauthorized
	}

	var claims user.Claims
	parsedToken, err := jwt.ParseWithClaims(token, &claims, func(_ *jwt.Token) (any, error) {
		return []byte(a.cfg.JWTSecret), nil
	}, jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return fiber.ErrUnauthorized
	}

	if !parsedToken.Valid {
		return fiber.ErrUnauthorized
	}

	fCtx.Locals(userKey, User{
		ID: claims.UserID,
	})
	return fCtx.Next()
}

func (a *auth) GenerateAndSetToken(fCtx fiber.Ctx, userModel *user.Model) error {
	expiresAt := time.Now().Add(time.Hour)
	token, err := a.generateJWT(userModel, expiresAt)
	if err != nil {
		return fmt.Errorf("generate JWT token: %w", err)
	}

	fCtx.Cookie(&fiber.Cookie{
		Name:     tokenCookie,
		Value:    token,
		HTTPOnly: true,
		Expires:  expiresAt,
	})

	return nil
}

func (a *auth) generateJWT(userModel *user.Model, expiresAt time.Time) (string, error) {
	now := jwt.NewNumericDate(time.Now())
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, user.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: now,
			IssuedAt:  now,
		},
		UserID: userModel.ID,
	})

	signedToken, err := token.SignedString([]byte(a.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signedToken, nil
}

func (a *auth) ClearToken(fCtx fiber.Ctx) {
	fCtx.Cookie(&fiber.Cookie{
		Name:     tokenCookie,
		HTTPOnly: true,
		Expires:  time.Now().Add(-time.Hour),
	})
}
