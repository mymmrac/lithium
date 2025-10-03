package project

import (
	"crypto/rand"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/mymmrac/lithium/pkg/module/action"
	"github.com/mymmrac/lithium/pkg/module/auth"
	"github.com/mymmrac/lithium/pkg/module/db"
	"github.com/mymmrac/lithium/pkg/module/id"
	"github.com/mymmrac/lithium/pkg/module/logger"
	"github.com/mymmrac/lithium/pkg/module/project"
	"github.com/mymmrac/lithium/pkg/module/storage"
	"github.com/mymmrac/lithium/pkg/module/user"
)

type handler struct {
	cfg               Config
	tx                db.Transaction
	userRepository    user.Repository
	projectRepository project.Repository
	actionCache       action.Cache
	actionRepository  action.Repository
	storage           storage.Storage
}

func RegisterHandlers(
	cfg Config, router fiber.Router, tx db.Transaction, userRepository user.Repository,
	projectRepository project.Repository, actionCache action.Cache, actionRepository action.Repository,
	storage storage.Storage,
) {
	h := &handler{
		cfg:               cfg,
		tx:                tx,
		userRepository:    userRepository,
		projectRepository: projectRepository,
		actionCache:       actionCache,
		actionRepository:  actionRepository,
		storage:           storage,
	}

	api := router.Group("/api/project", auth.RequireMiddleware)

	api.Get("/", h.getAllHandler)
	api.Post("/", h.createHandler)
	api.Get("/:projectID", h.getHandler)
	api.Put("/:projectID", h.updateHandler)
	api.Delete("/:projectID", h.deleteHandler)
}

type projectInfo struct {
	ID        id.ID  `json:"id"`
	Name      string `json:"name"`
	SubDomain string `json:"subDomain"`
}

func (h *handler) getAllHandler(fCtx fiber.Ctx) error {
	models, err := h.projectRepository.GetByOwnerID(fCtx, auth.MustUserFromContext(fCtx).ID)
	if err != nil {
		logger.FromContext(fCtx).Errorw("get projects", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	response := make([]projectInfo, len(models))
	for i, model := range models {
		response[i] = projectInfo{
			ID:        model.ID,
			Name:      model.Name,
			SubDomain: model.SubDomain,
		}
	}

	return fCtx.JSON(response)
}

func (h *handler) getHandler(fCtx fiber.Ctx) error {
	var request struct {
		ID id.ID `uri:"projectID" validate:"required"`
	}

	if err := fCtx.Bind().URI(&request); err != nil {
		logger.FromContext(fCtx).Warnw("get project, bad request", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}

	model, found, err := h.projectRepository.GetByID(fCtx, request.ID)
	if err != nil {
		logger.FromContext(fCtx).Errorw("get project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found || model.OwnerID != auth.MustUserFromContext(fCtx).ID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	return fCtx.JSON(&projectInfo{
		ID:        model.ID,
		Name:      model.Name,
		SubDomain: model.SubDomain,
	})
}

func (h *handler) createHandler(fCtx fiber.Ctx) error {
	var request struct {
		Name string `json:"name" validate:"alphanum_text,min=1,max=64"`
	}

	if err := fCtx.Bind().Body(&request); err != nil {
		logger.FromContext(fCtx).Warnw("create project, bad request", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}

	request.Name = strings.TrimSpace(request.Name)
	subDomainReplacer := strings.NewReplacer(" ", "-", "_", "-")
	subDomain := subDomainReplacer.Replace(strings.ToLower(request.Name)) + "-" + strings.ToLower(rand.Text()[:4])

	now := time.Now()
	err := h.projectRepository.Create(fCtx, &project.Model{
		ID:        id.New(),
		OwnerID:   auth.MustUserFromContext(fCtx).ID,
		Name:      request.Name,
		SubDomain: subDomain,
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
		ID   id.ID  `uri:"projectID" validate:"required"`
		Name string `json:"name"     validate:"alphanum_text,min=1,max=64"`
	}

	if err := fCtx.Bind().All(&request); err != nil {
		logger.FromContext(fCtx).Warnw("update project, bad request", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}

	model, found, err := h.projectRepository.GetByID(fCtx, request.ID)
	if err != nil {
		logger.FromContext(fCtx).Errorw("get user", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found || model.OwnerID != auth.MustUserFromContext(fCtx).ID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	err = h.projectRepository.UpdateName(fCtx, request.ID, request.Name)
	if err != nil {
		logger.FromContext(fCtx).Errorw("update project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.JSON(fiber.Map{"ok": true})
}

func (h *handler) deleteHandler(fCtx fiber.Ctx) error {
	var request struct {
		ID id.ID `uri:"projectID" validate:"required"`
	}

	if err := fCtx.Bind().URI(&request); err != nil {
		logger.FromContext(fCtx).Warnw("delete project, bad request", "error", err)
		return fiber.NewError(fiber.StatusBadRequest)
	}

	model, found, err := h.projectRepository.GetByID(fCtx, request.ID)
	if err != nil {
		logger.FromContext(fCtx).Errorw("get user", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found || model.OwnerID != auth.MustUserFromContext(fCtx).ID {
		return fiber.NewError(fiber.StatusNotFound)
	}

	ctx, err := h.tx.Begin(fCtx)
	if err != nil {
		logger.FromContext(fCtx).Errorw("begin transaction", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	defer func() { _ = h.tx.Rollback(ctx) }()

	actions, err := h.actionRepository.GetByProjectID(ctx, request.ID)
	if err != nil {
		logger.FromContext(ctx).Errorw("get project actions", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	for _, actionModel := range actions {
		if err = h.actionRepository.DeleteByID(ctx, actionModel.ID); err != nil {
			logger.FromContext(ctx).Errorw("delete action", "error", err)
			return fiber.NewError(fiber.StatusInternalServerError)
		}

		if actionModel.ModulePath != "" {
			if err = h.storage.Delete(ctx, h.cfg.ModuleBucket, actionModel.ModulePath); err != nil {
				logger.FromContext(ctx).Errorw("delete action module", "error", err)
				return fiber.NewError(fiber.StatusInternalServerError)
			}

			if err = h.actionCache.Remove(fCtx, actionModel.ID); err != nil {
				logger.FromContext(fCtx).Errorw("remove action from cache", "id", actionModel.ID, "error", err)
				return fiber.NewError(fiber.StatusInternalServerError)
			}
		}
	}

	if err = h.projectRepository.DeleteByID(ctx, request.ID); err != nil {
		logger.FromContext(ctx).Errorw("delete project", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	if err = h.tx.Commit(ctx); err != nil {
		logger.FromContext(fCtx).Errorw("commit transaction", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.JSON(fiber.Map{"ok": true})
}
