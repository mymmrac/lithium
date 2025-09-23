package auth

import (
	"crypto/rand"
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/mymmrac/lithium/pkg/module/id"
	"github.com/mymmrac/lithium/pkg/module/logger"
	"github.com/mymmrac/lithium/pkg/module/user"
)

type auth struct {
	userRepository user.Repository
}

func RegisterHandlers(router fiber.Router, userRepository user.Repository) error {
	a := &auth{
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

	fCtx.Cookie(&fiber.Cookie{
		Name:        tokenCookie,
		Value:       rand.Text() + "_" + userModel.ID.String(),
		HTTPOnly:    true,
		SessionOnly: true,
	})

	return fCtx.JSON(fiber.Map{"ok": true})
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

	err = a.userRepository.Create(fCtx, id.New(), request.Email, hashedPassword)
	if err != nil {
		if errors.Is(err, user.ErrAlreadyExists) {
			return fiber.NewError(fiber.StatusConflict, "user already exists")
		}

		logger.FromContext(fCtx).Errorw("create user", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.JSON(fiber.Map{"ok": true})
}

func (a *auth) logoutHandler(fCtx fiber.Ctx) error {
	fCtx.ClearCookie(tokenCookie)
	return fCtx.Redirect().To("/")
}
