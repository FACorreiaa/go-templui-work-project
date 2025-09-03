package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/FACorreiaa/go-templui/app/pages"
	"github.com/FACorreiaa/go-templui/assets"
	"github.com/FACorreiaa/go-templui/pkg/auth"
	"github.com/FACorreiaa/go-templui/pkg/logger"
	"github.com/a-h/templ"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Initialize logger
	err := logger.Init(zapcore.InfoLevel, 
		zap.String("service", "go-templui"),
		zap.String("version", "1.0.0"),
	)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	logger.Log.Info("Starting application")

	// Start pprof server on separate port
	go func() {
		logger.Log.Info("Starting pprof server on :6060")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			logger.Log.Error("Failed to start pprof server", zap.Error(err))
		}
	}()

	mux := http.NewServeMux()
	SetupAssetsRoutes(mux)

	// Initialize auth handlers
	authHandlers := auth.NewAuthHandlers()

	// Route handlers
	mux.Handle("GET /", templ.Handler(pages.Landing()))
	mux.Handle("GET /about", templ.Handler(pages.About()))
	mux.Handle("GET /projects", templ.Handler(pages.Projects()))
	
	// Auth routes
	mux.Handle("GET /login", templ.Handler(pages.LoginPage()))
	mux.Handle("GET /register", templ.Handler(pages.RegisterPage()))
	mux.Handle("GET /change-password", auth.JWTMiddleware(templ.Handler(pages.ChangePasswordPage())))
	
	// Error pages
	mux.Handle("GET /error/404", templ.Handler(pages.Error404Page()))
	mux.Handle("GET /error/500", templ.Handler(pages.Error500Page()))
	mux.Handle("GET /error/403", templ.Handler(pages.Error403Page()))
	mux.Handle("GET /error/401", templ.Handler(pages.ErrorAuthPage()))
	
	// Auth API endpoints
	mux.HandleFunc("POST /auth/login", authHandlers.LoginHandler)
	mux.HandleFunc("POST /auth/register", authHandlers.RegisterHandler)
	mux.HandleFunc("POST /auth/change-password", authHandlers.ChangePasswordHandler)
	mux.HandleFunc("GET /logout", authHandlers.LogoutHandler)

	// Start main server
	logger.Log.Info("Starting main server on :8090")
	if err := http.ListenAndServe(":8090", mux); err != nil {
		logger.Log.Fatal("Failed to start main server", zap.Error(err))
	}
}

func SetupAssetsRoutes(mux *http.ServeMux) {
	assetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Debug("Serving asset",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
		)
		// Always use embedded assets
		fs := http.FileServer(http.FS(assets.Assets))
		fs.ServeHTTP(w, r)
	})

	mux.Handle("GET /assets/", http.StripPrefix("/assets/", assetHandler))
}