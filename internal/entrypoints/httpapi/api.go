package httpapi

import (
	"context"
	"errors"
	"log"
	"nbox/internal/application"
	"nbox/internal/entrypoints/api/auth"
	"nbox/internal/entrypoints/api/handlers"
	"net/http"

	"github.com/norlis/httpgate/pkg/adapter/opa"

	_ "nbox/docs"

	"github.com/norlis/httpgate/pkg/adapter/apidriven/middleware"
	"github.com/norlis/httpgate/pkg/adapter/apidriven/presenters"
	"github.com/norlis/httpgate/pkg/application/health"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In
	Router *http.ServeMux
	Box    *handlers.BoxHandler
	Entry  *handlers.EntryHandler
	Static *handlers.StaticHandler
	Authn  *auth.Authn
	Status *health.Status
	Render presenters.Presenters
	Logger *zap.Logger
}

// NewHttpApi
// @title           nbox API
// @version         1.0
// @description     Esta es una API generada automáticamente con Swaggo.
// @termsOfService  http://swagger.io/terms/
// @contact.name   Norlis Viamonte
// @contact.url    http://www.example.com/support
// @contact.email  norlis.viamonte@gmail.com
// @host
// @BasePath  /
// @securityDefinitions.basic  BasicAuth
// @openapi 3.0.0
func NewHttpApi(params Params) {
	base := []middleware.Middleware{
		middleware.TraceId(middleware.WithHeaderName("TransactionId"), middleware.WithLogger(params.Logger)),
		middleware.APIErrorMiddleware(
			middleware.WithIntercept(http.StatusNotFound, http.StatusMethodNotAllowed, http.StatusInternalServerError),
			middleware.WithCustomMessage(http.StatusNotFound, "resource not found"),
			middleware.WithCustomMessage(http.StatusMethodNotAllowed, "method is not allowed for this resource."),
		),
		middleware.Recover(params.Logger, params.Render),
		middleware.RequestLogger(params.Logger),
		middleware.AllowAll(params.Logger).Middleware,
	}

	opaConfig := opa.Config{
		Query:        "data.authz.allow",
		PoliciesPath: "policies/authz", // Directorio con authz.rego
		DataFiles:    []string{},
	}

	authz, err := opa.NewOpaSdkClientFromConfig(context.Background(), opaConfig, params.Logger)

	if err != nil {
		log.Fatalf("No se pudo inicializar el cliente OPA: %v", err)
	}

	use := middleware.Chain(base...)

	params.Router.Handle("GET /status", use(params.Status))
	params.Router.Handle("GET /health", use(health.NewProbe(nil)))
	//router.Handle("GET /live", use(health.NewProbe(nil)))
	//router.Handle("GET /ready", use(health.NewProbe(nil)))

	params.Router.Handle("GET /swagger/", use(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	)))

	params.Router.Handle("POST /api/auth/token", use(params.Authn.TokenHandler()))

	api := http.NewServeMux()

	api.HandleFunc("POST /api/box", params.Box.UpsertBox)
	api.HandleFunc("GET /api/box", params.Box.List)
	api.HandleFunc("HEAD /api/box/{service}/{stage}/{template}", params.Box.Exist)
	api.HandleFunc("GET /api/box/{service}/{stage}/{template}", params.Box.Retrieve)
	api.HandleFunc("GET /api/box/{service}/{stage}/{template}/build", params.Box.Build)
	api.HandleFunc("GET /api/box/{service}/{stage}/{template}/vars", params.Box.ListVars)

	api.HandleFunc("POST /api/entry", params.Entry.Upsert)
	api.HandleFunc("GET /api/entry/key", params.Entry.GetByKey)
	api.HandleFunc("GET /api/entry/prefix", params.Entry.ListByPrefix)
	api.HandleFunc("DELETE /api/entry/key", params.Entry.DeleteKey)

	api.HandleFunc("GET /api/track/key", params.Entry.Tracking)

	api.HandleFunc("GET /api/static/environments", params.Static.Environments)

	useAuth := middleware.Chain(
		append(
			base,
			[]middleware.Middleware{
				params.Authn.Handler(),
				middleware.AuthorizationMiddleware(authz, FromContextExtractor),
			}...,
		)...,
	)
	params.Router.Handle("/api/", useAuth(api))
}

func FromContextExtractor(r *http.Request) (map[string]any, error) {
	user, ok := application.UserFromContext(r.Context())
	if !ok {
		return nil, errors.New("no user found in context")
	}

	return map[string]any{
		"roles": user.Roles,
	}, nil
}
