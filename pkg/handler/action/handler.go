package action

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/mymmrac/lithium/pkg/module/action"
	"github.com/mymmrac/lithium/pkg/module/auth"
	"github.com/mymmrac/lithium/pkg/module/id"
	"github.com/mymmrac/lithium/pkg/module/logger"
	"github.com/mymmrac/lithium/pkg/module/project"
)

type handler struct {
	actionRepository  action.Repository
	projectRepository project.Repository
}

func RegisterHandlers(router fiber.Router, actionRepository action.Repository, projectRepository project.Repository) {
	h := &handler{
		actionRepository:  actionRepository,
		projectRepository: projectRepository,
	}

	api := router.Group("/api/project/:projectID/action", auth.RequireMiddleware)

	api.Get("/", h.getAllHandler)
	api.Post("/", h.createHandler)
	api.Get("/:actionID", h.getHandler)
	api.Put("/:actionID", h.updateHandler)
	api.Delete("/:actionID", h.deleteHandler)
}

type actionInfo struct {
	ID      id.ID    `json:"id"`
	Name    string   `json:"name"`
	Path    string   `json:"path"`
	Methods []string `json:"methods"`
}

func (h *handler) getAllHandler(fCtx fiber.Ctx) error {
	var request struct {
		ProjectID id.ID `uri:"projectID" validate:"required"`
	}

	if err := fCtx.Bind().URI(&request); err != nil {
		logger.FromContext(fCtx).Warnw("get actions, bad request", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}

	projectModel, found, err := h.projectRepository.GetByID(fCtx, request.ProjectID)
	if err != nil {
		logger.FromContext(fCtx).Errorw("get project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found || projectModel.OwnerID != auth.MustUserFromContext(fCtx).ID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	models, err := h.actionRepository.GetByProjectID(fCtx, request.ProjectID)
	if err != nil {
		logger.FromContext(fCtx).Errorw("get actions", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	response := make([]actionInfo, len(models))
	for i, model := range models {
		response[i] = actionInfo{
			ID:      model.ID,
			Name:    model.Name,
			Path:    model.Path,
			Methods: model.Methods,
		}
	}

	return fCtx.JSON(response)
}

func (h *handler) getHandler(fCtx fiber.Ctx) error {
	var request struct {
		ProjectID id.ID `uri:"projectID" validate:"required"`
		ID        id.ID `uri:"actionID"  validate:"required"`
	}

	if err := fCtx.Bind().URI(&request); err != nil {
		logger.FromContext(fCtx).Warnw("get action, bad request", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}

	projectModel, found, err := h.projectRepository.GetByID(fCtx, request.ProjectID)
	if err != nil {
		logger.FromContext(fCtx).Errorw("get project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found || projectModel.OwnerID != auth.MustUserFromContext(fCtx).ID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	model, found, err := h.actionRepository.GetByID(fCtx, request.ID)
	if err != nil {
		logger.FromContext(fCtx).Errorw("get action", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound)
	}

	return fCtx.JSON(&actionInfo{
		ID:      model.ID,
		Name:    model.Name,
		Path:    model.Path,
		Methods: model.Methods,
	})
}

func (h *handler) createHandler(fCtx fiber.Ctx) error {
	var request struct {
		ProjectID id.ID    `uri:"projectID" validate:"required"`
		Name      string   `json:"name"     validate:"alphanum_text,min=1,max=64"`
		Path      string   `json:"path"     validate:"uri"`
		Methods   []string `json:"methods"  validate:"gt=0,unique,dive,oneof=GET POST PUT PATCH DELETE"`
	}

	if err := fCtx.Bind().All(&request); err != nil {
		logger.FromContext(fCtx).Warnw("create action, bad request", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}

	projectModel, found, err := h.projectRepository.GetByID(fCtx, request.ProjectID)
	if err != nil {
		logger.FromContext(fCtx).Errorw("get project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found || projectModel.OwnerID != auth.MustUserFromContext(fCtx).ID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	request.Name = strings.TrimSpace(request.Name)

	now := time.Now()
	err = h.actionRepository.Create(fCtx, &action.Model{
		ID:         id.New(),
		ProjectID:  request.ProjectID,
		Name:       request.Name,
		Path:       request.Path,
		Methods:    request.Methods,
		ModulePath: "",
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		logger.FromContext(fCtx).Errorw("create action", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.JSON(fiber.Map{"ok": true})
}

func (h *handler) updateHandler(fCtx fiber.Ctx) error {
	var request struct {
		ProjectID id.ID    `uri:"projectID" validate:"required"`
		ID        id.ID    `uri:"actionID"  validate:"required"`
		Name      string   `json:"name"     validate:"alphanum_text,min=1,max=64"`
		Path      string   `json:"path"     validate:"uri"`
		Methods   []string `json:"methods"  validate:"gt=0,unique,dive,oneof=GET POST PUT PATCH DELETE"`
	}

	if err := fCtx.Bind().All(&request); err != nil {
		logger.FromContext(fCtx).Warnw("update project, bad request", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}

	projectModel, found, err := h.projectRepository.GetByID(fCtx, request.ID)
	if err != nil {
		logger.FromContext(fCtx).Errorw("get project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found || projectModel.OwnerID != auth.MustUserFromContext(fCtx).ID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	request.Name = strings.TrimSpace(request.Name)

	err = h.actionRepository.UpdateInfo(fCtx, request.ID, request.Name, request.Path, request.Methods)
	if err != nil {
		logger.FromContext(fCtx).Errorw("update action", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.JSON(fiber.Map{"ok": true})
}

func (h *handler) deleteHandler(fCtx fiber.Ctx) error {
	var request struct {
		ID id.ID `uri:"actionID" validate:"required"`
	}

	if err := fCtx.Bind().URI(&request); err != nil {
		logger.FromContext(fCtx).Warnw("delete project, bad request", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}

	projectModel, found, err := h.projectRepository.GetByID(fCtx, request.ID)
	if err != nil {
		logger.FromContext(fCtx).Errorw("get project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found || projectModel.OwnerID != auth.MustUserFromContext(fCtx).ID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	// TODO: Remove module

	err = h.actionRepository.DeleteByID(fCtx, request.ID)
	if err != nil {
		logger.FromContext(fCtx).Errorw("delete action", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.JSON(fiber.Map{"ok": true})
}
