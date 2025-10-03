package action

import (
	"io"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/mymmrac/wape"

	"github.com/mymmrac/lithium/pkg/module/action"
	"github.com/mymmrac/lithium/pkg/module/auth"
	"github.com/mymmrac/lithium/pkg/module/db"
	"github.com/mymmrac/lithium/pkg/module/id"
	"github.com/mymmrac/lithium/pkg/module/logger"
	"github.com/mymmrac/lithium/pkg/module/project"
	"github.com/mymmrac/lithium/pkg/module/storage"
)

type handler struct {
	cfg               Config
	tx                db.Transaction
	actionCache       action.Cache
	actionRepository  action.Repository
	projectRepository project.Repository
	storage           storage.Storage
}

func RegisterHandlers(
	cfg Config, router fiber.Router, tx db.Transaction, actionCache action.Cache, actionRepository action.Repository,
	projectRepository project.Repository, storage storage.Storage,
) {
	h := &handler{
		cfg:               cfg,
		tx:                tx,
		actionCache:       actionCache,
		actionRepository:  actionRepository,
		projectRepository: projectRepository,
		storage:           storage,
	}

	api := router.Group("/api/project/:projectID/action", auth.RequireMiddleware)

	api.Get("/", h.getAllHandler)
	api.Post("/", h.createHandler)
	api.Post("/order", h.updateActionOrderHandler)
	api.Get("/:actionID", h.getHandler)
	api.Put("/:actionID", h.updateHandler)
	api.Post("/:actionID/upload", h.uploadHandler)
	api.Delete("/:actionID", h.deleteHandler)
}

type actionInfo struct {
	ID             id.ID    `json:"id"`
	Name           string   `json:"name"`
	Path           string   `json:"path"`
	Methods        []string `json:"methods"`
	ModuleUploaded bool     `json:"moduleUploaded"`
}

func (h *handler) getAllHandler(fCtx fiber.Ctx) error {
	var request struct {
		ProjectID id.ID `uri:"projectID" validate:"required"`
	}

	if err := fCtx.Bind().URI(&request); err != nil {
		logger.Warnw(fCtx, "get actions, bad request", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}

	projectModel, found, err := h.projectRepository.GetByID(fCtx, request.ProjectID)
	if err != nil {
		logger.Errorw(fCtx, "get project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found || projectModel.OwnerID != auth.MustUserFromContext(fCtx).ID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	models, err := h.actionRepository.GetByProjectID(fCtx, request.ProjectID)
	if err != nil {
		logger.Errorw(fCtx, "get actions", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	response := make([]actionInfo, len(models))
	for i, model := range models {
		response[i] = actionInfo{
			ID:             model.ID,
			Name:           model.Name,
			Path:           model.Path,
			Methods:        model.Methods,
			ModuleUploaded: model.ModulePath != "",
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
		logger.Warnw(fCtx, "get action, bad request", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}

	projectModel, found, err := h.projectRepository.GetByID(fCtx, request.ProjectID)
	if err != nil {
		logger.Errorw(fCtx, "get project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found || projectModel.OwnerID != auth.MustUserFromContext(fCtx).ID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	model, found, err := h.actionRepository.GetByID(fCtx, request.ID)
	if err != nil {
		logger.Errorw(fCtx, "get action", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found || model.ProjectID != request.ProjectID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	return fCtx.JSON(&actionInfo{
		ID:             model.ID,
		Name:           model.Name,
		Path:           model.Path,
		Methods:        model.Methods,
		ModuleUploaded: model.ModulePath != "",
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
		logger.Warnw(fCtx, "create action, bad request", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}

	projectModel, found, err := h.projectRepository.GetByID(fCtx, request.ProjectID)
	if err != nil {
		logger.Errorw(fCtx, "get project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found || projectModel.OwnerID != auth.MustUserFromContext(fCtx).ID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	count, err := h.actionRepository.CountByProjectID(fCtx, request.ProjectID)
	if err != nil {
		logger.Errorw(fCtx, "get actions count", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	request.Name = strings.TrimSpace(request.Name)

	now := time.Now()
	err = h.actionRepository.Create(fCtx, &action.Model{
		ID:         id.New(),
		ProjectID:  request.ProjectID,
		Name:       request.Name,
		Path:       request.Path,
		Methods:    request.Methods,
		Order:      count,
		ModulePath: "",
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		logger.Errorw(fCtx, "create action", "error", err)
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
		logger.Warnw(fCtx, "update action, bad request", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}

	projectModel, found, err := h.projectRepository.GetByID(fCtx, request.ProjectID)
	if err != nil {
		logger.Errorw(fCtx, "get project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found || projectModel.OwnerID != auth.MustUserFromContext(fCtx).ID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	model, found, err := h.actionRepository.GetByID(fCtx, request.ID)
	if err != nil {
		logger.Errorw(fCtx, "get action", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found || model.ProjectID != request.ProjectID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	request.Name = strings.TrimSpace(request.Name)

	err = h.actionRepository.UpdateInfo(fCtx, request.ID, request.Name, request.Path, request.Methods)
	if err != nil {
		logger.Errorw(fCtx, "update action", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.JSON(fiber.Map{"ok": true})
}

func (h *handler) uploadHandler(fCtx fiber.Ctx) error {
	var request struct {
		ProjectID id.ID `uri:"projectID" validate:"required"`
		ID        id.ID `uri:"actionID"  validate:"required"`
	}

	if err := fCtx.Bind().All(&request); err != nil {
		logger.Warnw(fCtx, "upload action, bad request", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}

	moduleFileHeader, err := fCtx.FormFile("module")
	if err != nil {
		logger.Warnw(fCtx, "upload action, bad request (file header)", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}

	moduleFile, err := moduleFileHeader.Open()
	if err != nil {
		logger.Warnw(fCtx, "upload action, bad request (open file)", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}
	defer func() { _ = moduleFile.Close() }()

	moduleData, err := io.ReadAll(moduleFile)
	if err != nil {
		logger.Errorw(fCtx, "upload action, read module file", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	projectModel, found, err := h.projectRepository.GetByID(fCtx, request.ProjectID)
	if err != nil {
		logger.Errorw(fCtx, "get project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	userID := auth.MustUserFromContext(fCtx).ID
	if !found || projectModel.OwnerID != userID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	env := wape.NewEnvironment()
	env.Modules = []wape.ModuleData{
		{
			Name: "main",
			Data: moduleData,
		},
	}

	// TODO: Take from config
	env.NetworkEnabled = true
	env.NetworksAllowAll = true
	env.NetworkAddressesAllowAll = true

	module, err := wape.NewPlugin(fCtx, env)
	if err != nil {
		logger.Warnw(fCtx, "instantiate module", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	if !module.FunctionExists("handler") {
		logger.Warnw(fCtx, "handler function doesn't exist")
		return fiber.NewError(fiber.StatusBadRequest, "Handler function is missing")
	}

	ctx, err := h.tx.Begin(fCtx)
	if err != nil {
		logger.Errorw(fCtx, "begin transaction", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	defer func() { _ = h.tx.Rollback(ctx) }()

	modulePath := path.Join(userID.String(), request.ProjectID.String(), request.ID.String()+".wasm")
	if err = h.actionRepository.UpdateModulePath(ctx, request.ID, modulePath); err != nil {
		logger.Errorw(fCtx, "update action module path", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	if err = h.storage.Upload(ctx, h.cfg.ModuleBucket, modulePath, moduleData); err != nil {
		logger.Errorw(fCtx, "upload action module", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	if err = h.tx.Commit(ctx); err != nil {
		logger.Errorw(fCtx, "commit transaction", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	if err = h.actionCache.Remove(fCtx, request.ID); err != nil {
		logger.Errorw(fCtx, "remove action from cache", "id", request.ID, "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.JSON(fiber.Map{"ok": true})
}

func (h *handler) updateActionOrderHandler(fCtx fiber.Ctx) error {
	var request struct {
		ProjectID id.ID   `uri:"projectID" validate:"required"`
		IDs       []id.ID `json:"ids"      validate:"gt=0,dive,required"`
	}

	if err := fCtx.Bind().All(&request); err != nil {
		logger.Warnw(fCtx, "update action order, bad request", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}

	projectModel, found, err := h.projectRepository.GetByID(fCtx, request.ProjectID)
	if err != nil {
		logger.Errorw(fCtx, "get project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found || projectModel.OwnerID != auth.MustUserFromContext(fCtx).ID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	models, err := h.actionRepository.GetByProjectID(fCtx, request.ProjectID)
	if err != nil {
		logger.Errorw(fCtx, "get actions", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	if len(models) != len(request.IDs) {
		logger.Warnw(fCtx, "update action order, length doesn't match",
			"expected", len(models), "actual", len(request.IDs),
		)
		return fiber.NewError(fiber.StatusBadRequest)
	}
	for _, modelID := range request.IDs {
		if !slices.ContainsFunc(models, func(model action.Model) bool {
			return model.ID == modelID
		}) {
			logger.Warnw(fCtx, "update action order, unexpected action", "id", modelID)
			return fiber.NewError(fiber.StatusBadRequest)
		}
	}

	err = h.actionRepository.UpdateOrder(fCtx, request.IDs)
	if err != nil {
		logger.Errorw(fCtx, "update action order", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.JSON(fiber.Map{"ok": true})
}

func (h *handler) deleteHandler(fCtx fiber.Ctx) error {
	var request struct {
		ProjectID id.ID `uri:"projectID" validate:"required"`
		ID        id.ID `uri:"actionID"  validate:"required"`
	}

	if err := fCtx.Bind().URI(&request); err != nil {
		logger.Warnw(fCtx, "delete action, bad request", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}

	projectModel, found, err := h.projectRepository.GetByID(fCtx, request.ProjectID)
	if err != nil {
		logger.Errorw(fCtx, "get project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found || projectModel.OwnerID != auth.MustUserFromContext(fCtx).ID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	model, found, err := h.actionRepository.GetByID(fCtx, request.ID)
	if err != nil {
		logger.Errorw(fCtx, "get action", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found || model.ProjectID != request.ProjectID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	ctx, err := h.tx.Begin(fCtx)
	if err != nil {
		logger.Errorw(fCtx, "begin transaction", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	defer func() { _ = h.tx.Rollback(ctx) }()

	if err = h.actionRepository.DeleteByID(ctx, model.ID); err != nil {
		logger.Errorw(fCtx, "delete action", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	if model.ModulePath != "" {
		if err = h.storage.Delete(ctx, h.cfg.ModuleBucket, model.ModulePath); err != nil {
			logger.Errorw(fCtx, "delete action module", "error", err)
			return fiber.NewError(fiber.StatusInternalServerError)
		}

		if err = h.actionCache.Remove(fCtx, request.ID); err != nil {
			logger.Errorw(fCtx, "remove action from cache", "id", request.ID, "error", err)
			return fiber.NewError(fiber.StatusInternalServerError)
		}
	}

	if err = h.tx.Commit(ctx); err != nil {
		logger.Errorw(fCtx, "commit transaction", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.JSON(fiber.Map{"ok": true})
}
