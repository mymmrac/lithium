package auth

import (
	"crypto/rand"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

type auth struct{}

func RegisterHandlers(router fiber.Router) error {
	a := &auth{}

	router.Post("/login", a.loginHandler)
	router.Post("/register", a.registerHandler)
	router.Post("/logout", a.logoutHandler)

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

	fmt.Println(request)

	fCtx.Cookie(&fiber.Cookie{
		Name:        tokenCookie,
		Value:       rand.Text(),
		HTTPOnly:    true,
		SessionOnly: true,
	})

	return fCtx.JSON(fiber.Map{"ok": true})
}

func (a *auth) registerHandler(fCtx fiber.Ctx) error {
	return nil
}

func (a *auth) logoutHandler(fCtx fiber.Ctx) error {
	fCtx.ClearCookie(tokenCookie)
	return fCtx.Redirect().To("/")
}
