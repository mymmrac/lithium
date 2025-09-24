package project

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/mymmrac/lithium/pkg/module/auth"
	"github.com/mymmrac/lithium/pkg/module/id"
	"github.com/mymmrac/lithium/pkg/module/logger"
	"github.com/mymmrac/lithium/pkg/module/project"
)

type handler struct {
	projectRepository project.Repository
}

func RegisterHandlers(router fiber.Router, projectRepository project.Repository) {
	h := &handler{
		projectRepository: projectRepository,
	}

	api := router.Group("/api/project", auth.RequireMiddleware)

	api.Post("/", h.createHandler)
	api.Put("/", h.updateHandler)
	api.Delete("/", h.deleteHandler)
}

func (h *handler) createHandler(fCtx fiber.Ctx) error {
	var request struct {
		Name string `validate:"min=1,max=64"`
	}

	if err := fCtx.Bind().Body(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest)
	}

	now := time.Now()
	err := h.projectRepository.Create(fCtx, &project.Model{
		ID:        id.New(),
		OwnerID:   auth.MustUserFromContext(fCtx).ID,
		Name:      request.Name,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		logger.FromContext(fCtx).Errorw("create project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.JSON(fiber.Map{"ok": true})
}

func (h *handler) updateHandler(fCtx fiber.Ctx) error {
	var request struct {
		ID   id.ID  `validate:"required"`
		Name string `validate:"min=1,max=64"`
	}

	if err := fCtx.Bind().Body(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest)
	}

	err := h.projectRepository.UpdateName(fCtx, request.ID, auth.MustUserFromContext(fCtx).ID, request.Name)
	if err != nil {
		logger.FromContext(fCtx).Errorw("update project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.JSON(fiber.Map{"ok": true})
}

func (h *handler) deleteHandler(fCtx fiber.Ctx) error {
	var request struct {
		ID id.ID `validate:"required"`
	}

	if err := fCtx.Bind().Body(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest)
	}

	err := h.projectRepository.DeleteByID(fCtx, request.ID, auth.MustUserFromContext(fCtx).ID)
	if err != nil {
		logger.FromContext(fCtx).Errorw("delete project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.JSON(fiber.Map{"ok": true})
}
