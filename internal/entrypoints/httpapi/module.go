package httpapi

import (
	"context"
	"errors"
	"fmt"
	"nbox/internal/application"
	"net"
	"net/http"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewHttpServerMux(lc fx.Lifecycle, logger *zap.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	listener := net.JoinHostPort(application.Address, application.Port)

	server := &http.Server{
		Addr:              listener,
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info(
				"ListenAndServe",
				zap.String("addr", net.JoinHostPort(application.Address, application.Port)),
			)
			go func() {
				if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Error("Error al iniciar servidor HTTP: %v", zap.Error(err))
				}
			}()
			logger.Info(fmt.Sprintf("Servidor HTTP escuchando en %s", listener))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Deteniendo servidor HTTP...")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := server.Shutdown(shutdownCtx); err != nil {
				logger.Error("Error durante el apagado del servidor HTTP: %v", zap.Error(err))
				return err
			}
			logger.Info("Servidor HTTP detenido correctamente.")
			return nil
		},
	})

	return mux
}
