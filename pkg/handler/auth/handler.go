package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/mymmrac/lithium/pkg/module/id"
	"github.com/mymmrac/lithium/pkg/module/logger"
	"github.com/mymmrac/lithium/pkg/module/user"
)

type auth struct {
	cfg            Config
	userRepository user.Repository
}

func RegisterHandlers(cfg Config, router fiber.Router, userRepository user.Repository) error {
	a := &auth{
		cfg:            cfg,
		userRepository: userRepository,
	}

	api := router.Group("/api")

	api.Post("/login", a.loginHandler)
	api.Post("/register", a.registerHandler)
	api.Post("/logout", a.logoutHandler) // TODO: Add auth middleware

	return nil
}

const tokenCookie = "token"

func (a *auth) loginHandler(fCtx fiber.Ctx) error {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := fCtx.Bind().Body(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest)
	}

	userModel, found, err := a.userRepository.GetByEmail(fCtx, request.Email)
	if err != nil {
		logger.FromContext(fCtx).Errorw("get user by email", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found {
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	match, needsRehash, err := user.ComparePasswordAndHash(request.Password, userModel.Password)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !match {
		return fiber.NewError(fiber.StatusUnauthorized)
	}
	if needsRehash {
		logger.FromContext(fCtx).Warnw("user needs to rehash password", "user-id", userModel.ID)
	}

	if err = a.setTokenCookie(fCtx, userModel); err != nil {
		logger.FromContext(fCtx).Errorw("set token cookie", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.JSON(fiber.Map{"ok": true})
}

func (a *auth) setTokenCookie(fCtx fiber.Ctx, userModel *user.Model) error {
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

func (a *auth) registerHandler(fCtx fiber.Ctx) error {
	var request struct {
		Email    string `json:"email"    validate:"email"`
		Password string `json:"password" validate:"min=8,max=64"`
	}

	if err := fCtx.Bind().Body(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest)
	}

	hashedPassword, err := user.HashPassword(request.Password)
	if err != nil {
		logger.FromContext(fCtx).Errorw("hash password", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	userModel := &user.Model{
		ID:       id.New(),
		Email:    request.Email,
		Password: hashedPassword,
	}

	err = a.userRepository.Create(fCtx, userModel)
	if err != nil {
		if errors.Is(err, user.ErrAlreadyExists) {
			return fiber.NewError(fiber.StatusConflict, "user already exists")
		}

		logger.FromContext(fCtx).Errorw("create user", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	if err = a.setTokenCookie(fCtx, userModel); err != nil {
		logger.FromContext(fCtx).Errorw("set token cookie", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.JSON(fiber.Map{"ok": true})
}

func (a *auth) logoutHandler(fCtx fiber.Ctx) error {
	fCtx.Cookie(&fiber.Cookie{
		Name:     tokenCookie,
		HTTPOnly: true,
		Expires:  time.Now().Add(-1 * time.Hour),
	})
	return fCtx.Redirect().To("/")
}
