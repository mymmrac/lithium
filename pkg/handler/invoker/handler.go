package invoker

import (
	"github.com/gofiber/fiber/v3"

	"github.com/mymmrac/lithium/pkg/module/action"
	"github.com/mymmrac/lithium/pkg/module/logger"
	"github.com/mymmrac/lithium/pkg/module/project"
)

type Invoker interface {
	Middleware(fCtx fiber.Ctx) error
}

type invoker struct {
	actionRepository  action.Repository
	projectRepository project.Repository
}

func NewInvoker(actionRepository action.Repository, projectRepository project.Repository) Invoker {
	return &invoker{
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
			// TODO: Actual implementation
			return fCtx.SendString(actionModel.ID.String())
		})
	}
	app.Handler()(fCtx.RequestCtx())

	return nil
}
