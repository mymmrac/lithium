package auth

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"

	authm "github.com/mymmrac/lithium/pkg/module/auth"
	"github.com/mymmrac/lithium/pkg/module/id"
	"github.com/mymmrac/lithium/pkg/module/logger"
	"github.com/mymmrac/lithium/pkg/module/user"
)

type handler struct {
	auth           authm.Auth
	userRepository user.Repository
}

func RegisterHandlers(router fiber.Router, auth authm.Auth, userRepository user.Repository) error {
	h := &handler{
		auth:           auth,
		userRepository: userRepository,
	}

	api := router.Group("/api")

	api.Post("/login", h.loginHandler)
	api.Post("/register", h.registerHandler)
	api.Post("/logout", authm.RequireMiddleware, h.logoutHandler)

	return nil
}

func (h *handler) loginHandler(fCtx fiber.Ctx) error {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := fCtx.Bind().Body(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest)
	}

	userModel, found, err := h.userRepository.GetByEmail(fCtx, request.Email)
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

	if err = h.auth.GenerateAndSetToken(fCtx, userModel); err != nil {
		logger.FromContext(fCtx).Errorw("set token cookie", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.JSON(fiber.Map{"ok": true})
}

func (h *handler) registerHandler(fCtx fiber.Ctx) error {
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

	now := time.Now()
	userModel := &user.Model{
		ID:        id.New(),
		Email:     request.Email,
		Password:  hashedPassword,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = h.userRepository.Create(fCtx, userModel)
	if err != nil {
		if errors.Is(err, user.ErrAlreadyExists) {
			return fiber.NewError(fiber.StatusConflict, "Email already used")
		}

		logger.FromContext(fCtx).Errorw("create user", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	if err = h.auth.GenerateAndSetToken(fCtx, userModel); err != nil {
		logger.FromContext(fCtx).Errorw("set token cookie", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.JSON(fiber.Map{"ok": true})
}

func (h *handler) logoutHandler(fCtx fiber.Ctx) error {
	h.auth.ClearToken(fCtx)
	return fCtx.Redirect().To("/")
}
