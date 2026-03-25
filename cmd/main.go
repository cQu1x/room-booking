// @title           Booking Service API
// @version         1.0
// @description     Meeting room booking service for Avito internship.
// @host            localhost:8080
// @BasePath        /
//
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     JWT token obtained from /dummyLogin, /login, or /register. Format: "Bearer <token>"
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/avito-internships/test-backend-1-cQu1x/docs"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/config"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/handler"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/infrastructure/jwt"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/infrastructure/postgres"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/usecase"
	bookingusecase "github.com/avito-internships/test-backend-1-cQu1x/internal/usecase/booking"
	roomusecase "github.com/avito-internships/test-backend-1-cQu1x/internal/usecase/room"
	scheduleusecase "github.com/avito-internships/test-backend-1-cQu1x/internal/usecase/schedule"
	slotusecase "github.com/avito-internships/test-backend-1-cQu1x/internal/usecase/slot"
	"github.com/avito-internships/test-backend-1-cQu1x/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.LoadConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	// ── Database ──────────────────────────────────────────────────────────────
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	pool, err := pgxpool.New(dbCtx, cfg.DB.DSN())
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}

	if err := pool.Ping(dbCtx); err != nil {
		pool.Close()
		log.Fatalf("ping database: %v", err)
	}

	// ── Migrations ────────────────────────────────────────────────────────────
	if _, err := pool.Exec(dbCtx, migrations.InitSQL); err != nil {
		pool.Close()
		log.Fatalf("run migrations: %v", err)
	}
	defer pool.Close()

	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo := postgres.NewUserRepository(pool)
	roomRepo := postgres.NewRoomRepository(pool)
	scheduleRepo := postgres.NewScheduleRepository(pool)
	slotRepo := postgres.NewSlotRepository(pool)
	bookingRepo := postgres.NewBookingRepository(pool)

	// ── Services ──────────────────────────────────────────────────────────────
	tokenManager := jwt.NewTokenManager(cfg.JWT.Secret)

	authSvc := usecase.NewAuthUseCase(userRepo, tokenManager)
	roomSvc := roomusecase.NewService(roomRepo)
	slotSvc := slotusecase.NewService(slotRepo, scheduleRepo)
	scheduleSvc := scheduleusecase.NewService(scheduleRepo, roomRepo, slotSvc)
	bookingSvc := bookingusecase.NewService(bookingRepo, slotRepo)

	// ── Handlers ──────────────────────────────────────────────────────────────
	handlers := handler.Handlers{
		Auth:     handler.NewAuthHandler(authSvc),
		Room:     handler.NewRoomHandler(roomSvc),
		Schedule: handler.NewScheduleHandler(scheduleSvc),
		Slot:     handler.NewSlotHandler(slotSvc, roomSvc),
		Booking:  handler.NewBookingHandler(bookingSvc),
	}

	router := handler.NewRouter(handlers, tokenManager)

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("shutting down server...")
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutCancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}()

	log.Printf("starting server on :%s", cfg.App.Port)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server: %v", err)
	}
	log.Println("server stopped")
}
