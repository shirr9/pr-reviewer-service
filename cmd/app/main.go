package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"time"

	"github.com/go-playground/validator/v10"
	"github.com/shirr9/pr-reviewer-service/internal/app/config"
	"github.com/shirr9/pr-reviewer-service/internal/app/handler"
	"github.com/shirr9/pr-reviewer-service/internal/app/service"
	"github.com/shirr9/pr-reviewer-service/internal/infrastructure/logger"
	"github.com/shirr9/pr-reviewer-service/internal/infrastructure/persistence/postgres"
)

func main() {
	cfg, err := config.MustLoad("configs/config.yml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	appLogger := logger.NewLogger(cfg.Server.Env, os.Stdout)
	appLogger.Info("starting pr-reviewer-service", "env", cfg.Server.Env)

	ctx := context.Background()
	storage, err := postgres.NewStorage(ctx, *cfg)
	if err != nil {
		appLogger.Error("failed to connect to database", "error", err)
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer storage.Close()
	appLogger.Info("connected to database")

	prRepo := storage.NewPullRequestRepository()
	reviewerRepo := storage.NewReviewerRepository()
	userRepo := storage.NewUserRepository()
	teamRepo := storage.NewTeamRepository()
	uow := storage.NewUnitOfWork()

	prService := service.NewPullRequestService(prRepo, reviewerRepo, userRepo, uow, appLogger)
	userService := service.NewUserService(userRepo, prRepo, appLogger)
	teamService := service.NewTeamService(teamRepo, userRepo, prRepo, reviewerRepo, uow, appLogger)
	statisticsService := service.NewStatisticsService(userRepo, prRepo, reviewerRepo, appLogger)

	validate := validator.New()

	prHandler := handler.NewPullRequestHandler(prService, appLogger, validate)
	userHandler := handler.NewUserHandler(userService, appLogger, validate)
	teamHandler := handler.NewTeamHandler(teamService, appLogger, validate)
	statisticsHandler := handler.NewStatisticsHandler(statisticsService, appLogger)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /team/add", teamHandler.AddTeam)
	mux.HandleFunc("GET /team/get", teamHandler.GetTeam)
	mux.HandleFunc("POST /team/deactivate", teamHandler.DeactivateTeam)
	mux.HandleFunc("POST /users/setIsActive", userHandler.SetIsActive)
	mux.HandleFunc("GET /users/getReview", userHandler.GetReview)
	mux.HandleFunc("POST /pullRequest/create", prHandler.CreatePR)
	mux.HandleFunc("POST /pullRequest/merge", prHandler.MergePR)
	mux.HandleFunc("POST /pullRequest/reassign", prHandler.ReassignReviewer)
	mux.HandleFunc("GET /statistics", statisticsHandler.GetStatistics)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		appLogger.Info("starting HTTP server", "addr", addr)
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("server failed", "error", err)
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	appLogger.Info("shutdown server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		appLogger.Error("server shutdown failed", "error", err)
		log.Fatal("server shutdown:", err)
	}

	select {
	case <-ctx.Done():
		appLogger.Info("timeout of 5 seconds.")
	default:
		appLogger.Info("server exited gracefully")
	}

	log.Println("server exiting")
}
