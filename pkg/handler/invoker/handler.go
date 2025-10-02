package invoker

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/mymmrac/lithium/pkg/module/action"
	"github.com/mymmrac/lithium/pkg/module/logger"
	"github.com/mymmrac/lithium/pkg/module/project"
	"github.com/mymmrac/lithium/pkg/module/storage"
)

type Invoker interface {
	Middleware(fCtx fiber.Ctx) error
}

type invoker struct {
	cfg               Config
	storage           storage.Storage
	actionRepository  action.Repository
	projectRepository project.Repository
}

func NewInvoker(
	cfg Config, storage storage.Storage, actionRepository action.Repository, projectRepository project.Repository,
) Invoker {
	return &invoker{
		cfg:               cfg,
		storage:           storage,
		actionRepository:  actionRepository,
		projectRepository: projectRepository,
	}
}

func (i *invoker) Middleware(fCtx fiber.Ctx) error {
	subDomains := fCtx.Subdomains()
	if len(subDomains) == 0 || len(subDomains) > 1 {
		return fCtx.Next()
	}

	subDomain := subDomains[0]
	if len(subDomain) == 0 {
		return fCtx.Next()
	}

	return i.invoke(fCtx, subDomain)
}

func (i *invoker) invoke(fCtx fiber.Ctx, subDomain string) error {
	projectModel, found, err := i.projectRepository.GetBySubDomain(fCtx, subDomain)
	if err != nil {
		logger.FromContext(fCtx).Errorw("get project by subdomain", "sub-domain", subDomain, "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound)
	}

	actions, err := i.actionRepository.GetByProjectID(fCtx, projectModel.ID)
	if err != nil {
		logger.FromContext(fCtx).Errorw("get actions by project", "project-id", projectModel.ID, "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if len(actions) == 0 {
		return fiber.NewError(fiber.StatusNotFound)
	}

	app := fiber.New()
	for _, actionModel := range actions {
		app.Route(actionModel.Path).Add(actionModel.Methods, func(fCtx fiber.Ctx) error {
			return i.invokeAction(fCtx, actionModel)
		})
	}
	app.Handler()(fCtx.RequestCtx())

	return nil
}

func (i *invoker) invokeAction(fCtx fiber.Ctx, action action.Model) error {
	if action.ModulePath == "" {
		return fiber.NewError(fiber.StatusNotImplemented)
	}

	moduleData, err := i.storage.Download(fCtx, i.cfg.ModuleBucket, action.ModulePath)
	if err != nil {
		logger.FromContext(fCtx).Errorw("download module", "module", action.ModulePath, "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return fCtx.SendString(strconv.Itoa(len(moduleData)))
}
