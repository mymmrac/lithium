package invoker

import (
	extism "github.com/extism/go-sdk"
	"github.com/gofiber/fiber/v3"
	"github.com/mymmrac/wape"

	"github.com/mymmrac/lithium/pkg/module/action"
	"github.com/mymmrac/lithium/pkg/module/logger"
	"github.com/mymmrac/lithium/pkg/module/project"
	"github.com/mymmrac/lithium/pkg/module/protocol"
	"github.com/mymmrac/lithium/pkg/module/storage"
)

type Invoker interface {
	Middleware(fCtx fiber.Ctx) error
}

type invoker struct {
	cfg               Config
	storage           storage.Storage
	actionCache       action.Cache
	actionRepository  action.Repository
	projectRepository project.Repository
}

func NewInvoker(
	cfg Config, storage storage.Storage, actionCache action.Cache, actionRepository action.Repository,
	projectRepository project.Repository,
) Invoker {
	return &invoker{
		cfg:               cfg,
		storage:           storage,
		actionCache:       actionCache,
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
		logger.Errorw(fCtx, "get project by subdomain", "sub-domain", subDomain, "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound)
	}

	actions, err := i.actionRepository.GetByProjectID(fCtx, projectModel.ID)
	if err != nil {
		logger.Errorw(fCtx, "get actions by project", "project-id", projectModel.ID, "error", err)
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

	module, ok, err := i.actionCache.Get(fCtx, action.ID)
	if err != nil {
		logger.Errorw(fCtx, "get action module from cache", "id", action.ID, "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if !ok {
		var moduleData []byte
		moduleData, err = i.storage.Download(fCtx, i.cfg.ModuleBucket, action.ModulePath)
		if err != nil {
			logger.Errorw(fCtx, "download module", "module", action.ModulePath, "error", err)
			return fiber.NewError(fiber.StatusInternalServerError)
		}

		env := wape.NewEnvironment()
		env.Modules = []wape.ModuleData{
			{
				Name: "main",
				Data: moduleData,
			},
		}

		var compiledPlugin *extism.CompiledPlugin
		compiledPlugin, err = wape.NewCompiledPlugin(fCtx, env)
		if err != nil {
			logger.Errorw(fCtx, "compile module", "error", err)
			return fiber.NewError(fiber.StatusInternalServerError)
		}

		module.CompiledPlugin = compiledPlugin
		module.PluginInstanceConfig = env.MakePluginInstanceConfig()

		if err = i.actionCache.Set(fCtx, action.ID, module); err != nil {
			logger.Warnw(fCtx, "set action module cache", "error", err)
		}
	}

	plugin, err := module.CompiledPlugin.Instance(fCtx, module.PluginInstanceConfig)
	if err != nil {
		logger.Errorw(fCtx, "instantiate module", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	request, err := (&protocol.Request{
		URL:     string(fCtx.Request().URI().FullURI()),
		Method:  fCtx.Method(),
		Headers: fCtx.GetHeaders(),
		Body:    string(fCtx.Body()),
	}).Marshal()
	if err != nil {
		logger.Errorw(fCtx, "marshal request", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	exitCode, responseData, err := plugin.CallWithContext(fCtx, "handler", request)
	if err != nil {
		logger.Warnw(fCtx, "call module", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if exitCode != 0 {
		logger.Warnw(fCtx, "call module", "exit-code", exitCode)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	var response protocol.Response
	if err = response.Unmarshal(responseData); err != nil {
		logger.Warnw(fCtx, "unmarshal response", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}
	if response.StatusCode == 0 {
		response.StatusCode = fiber.StatusOK
	}

	resp := fCtx.Response()
	resp.SetStatusCode(response.StatusCode)
	for key, values := range response.Headers {
		for _, value := range values {
			resp.Header.Add(key, value)
		}
	}
	resp.SetBodyString(response.Body)

	return nil
}
