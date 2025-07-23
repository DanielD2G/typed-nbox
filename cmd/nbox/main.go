package main

import (
	"flag"
	"fmt"
	"log"
	"nbox/internal/adapters/aws"
	"nbox/internal/adapters/persistence"
	"nbox/internal/application"
	"nbox/internal/domain"
	"nbox/internal/entrypoints/api/auth"
	"nbox/internal/entrypoints/api/handlers"
	"nbox/internal/entrypoints/httpapi"
	"nbox/internal/usecases"
	"nbox/pkg/logger"
	"os"

	"go.uber.org/zap"

	"github.com/norlis/httpgate/pkg/adapter/apidriven/presenters"
	status "github.com/norlis/httpgate/pkg/application/health"
	"go.uber.org/fx"
)

// banner
// https://patorjk.com/software/taag/#p=display&f=Doh&t=NBOX
const banner = `
                                                                                    
NNNNNNNN        NNNNNNNNBBBBBBBBBBBBBBBBB        OOOOOOOOO     XXXXXXX       XXXXXXX
N:::::::N       N::::::NB::::::::::::::::B     OO:::::::::OO   X:::::X       X:::::X
N::::::::N      N::::::NB::::::BBBBBB:::::B  OO:::::::::::::OO X:::::X       X:::::X
N:::::::::N     N::::::NBB:::::B     B:::::BO:::::::OOO:::::::OX::::::X     X::::::X
N::::::::::N    N::::::N  B::::B     B:::::BO::::::O   O::::::OXXX:::::X   X:::::XXX
N:::::::::::N   N::::::N  B::::B     B:::::BO:::::O     O:::::O   X:::::X X:::::X   
N:::::::N::::N  N::::::N  B::::BBBBBB:::::B O:::::O     O:::::O    X:::::X:::::X    
N::::::N N::::N N::::::N  B:::::::::::::BB  O:::::O     O:::::O     X:::::::::X     
N::::::N  N::::N:::::::N  B::::BBBBBB:::::B O:::::O     O:::::O     X:::::::::X     
N::::::N   N:::::::::::N  B::::B     B:::::BO:::::O     O:::::O    X:::::X:::::X    
N::::::N    N::::::::::N  B::::B     B:::::BO:::::O     O:::::O   X:::::X X:::::X   
N::::::N     N:::::::::N  B::::B     B:::::BO::::::O   O::::::OXXX:::::X   X:::::XXX
N::::::N      N::::::::NBB:::::BBBBBB::::::BO:::::::OOO:::::::OX::::::X     X::::::X
N::::::N       N:::::::NB:::::::::::::::::B  OO:::::::::::::OO X:::::X       X:::::X
N::::::N        N::::::NB::::::::::::::::B     OO:::::::::OO   X:::::X       X:::::X
NNNNNNNN         NNNNNNNBBBBBBBBBBBBBBBBB        OOOOOOOOO     XXXXXXX       XXXXXXX
                                                                                    
`

func main() {

	fmt.Print(banner)

	flag.StringVar(&application.Port, "port", "7337", "--port=7337")
	flag.StringVar(&application.Address, "address", "", "--address=0.0.0.0")
	flag.Parse()

	app := fx.New(
		fx.Provide(logger.NewLogger),
		fx.Provide(aws.NewAwsConfig),
		fx.Provide(aws.NewS3Client),
		fx.Provide(aws.NewDynamodbClient),
		fx.Provide(aws.NewSsmClient),
		fx.Provide(aws.NewS3TemplateStore),
		fx.Provide(aws.NewDynamodbBackend),
		fx.Provide(aws.NewSecureParameterStore),
		fx.Provide(handlers.NewEntryHandler),
		fx.Provide(handlers.NewBoxHandler),
		fx.Provide(handlers.NewStaticHandler),
		fx.Provide(usecases.NewPathUseCase),
		fx.Provide(usecases.NewEntryUseCase),
		fx.Provide(usecases.NewBox),
		fx.Provide(application.NewConfigFromEnv),
		fx.Provide(func() *status.Status {
			version := application.GitHash
			if version == "" {
				version = "unknown.dev"
			}
			return status.NewStatus(version)
		}),
		fx.Provide(presenters.NewPresenters),
		fx.Provide(httpapi.NewHttpServerMux),
		fx.Provide(func() domain.UserRepository {
			credentials := os.Getenv(application.EnvCredentials)
			repo, err := persistence.NewInMemoryUserRepository([]byte(credentials))
			if err != nil {
				log.Fatal(err)
			}
			return repo
		}),
		fx.Provide(func(config *application.Config, render presenters.Presenters, logger *zap.Logger, repo domain.UserRepository) *auth.Authn {
			return auth.NewAuthn(application.EnvCredentials, config, render, logger, repo)
		}),
		fx.Invoke(httpapi.NewHttpApi),
	)

	if err := app.Err(); err != nil {
		log.Panicf("Error en la inicialización de la aplicación FX: %v\n", err)
	}

	app.Run()

}
