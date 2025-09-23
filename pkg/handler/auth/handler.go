package auth

import (
	"crypto/rand"

	"github.com/gofiber/fiber/v3"
)

type auth struct{}

func RegisterHandlers(router fiber.Router) error {
	a := &auth{}
	api := router.Group("/api")

	api.Post("/login", a.loginHandler)
	api.Post("/register", a.registerHandler)
	api.Post("/logout", a.logoutHandler)

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

	// fmt.Println(request)

	fCtx.Cookie(&fiber.Cookie{
		Name:        tokenCookie,
		Value:       rand.Text(),
		HTTPOnly:    true,
		SessionOnly: true,
	})

	return fCtx.JSON(fiber.Map{"ok": true})
}

func (a *auth) registerHandler(fCtx fiber.Ctx) error {
	_ = fCtx
	return nil
}

func (a *auth) logoutHandler(fCtx fiber.Ctx) error {
	fCtx.ClearCookie(tokenCookie)
	return fCtx.Redirect().To("/")
}
