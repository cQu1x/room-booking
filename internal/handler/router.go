package handler

import (
	"net/http"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	jwtpkg "github.com/avito-internships/test-backend-1-cQu1x/internal/infrastructure/jwt"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Handlers struct {
	Auth     *AuthHandler
	Room     *RoomHandler
	Schedule *ScheduleHandler
	Slot     *SlotHandler
	Booking  *BookingHandler
}

// NewRouter собирает HTTP-маршрутизатор и подключает все обработчики с нужными middleware.
func NewRouter(h Handlers, tokenManager *jwtpkg.TokenManager) http.Handler {
	mux := http.NewServeMux()

	auth := func(next http.Handler) http.Handler {
		return AuthMiddleware(tokenManager, next)
	}
	adminOnly := RequireRole(entity.RoleAdmin)
	userOnly := RequireRole(entity.RoleUser)
	anyRole := RequireRole(entity.RoleAdmin, entity.RoleUser)

	// ── Public ────────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /_info", info)
	mux.HandleFunc("POST /dummyLogin", h.Auth.DummyLogin)
	mux.HandleFunc("POST /register", h.Auth.Register)
	mux.HandleFunc("POST /login", h.Auth.Login)

	// ── Swagger UI ────────────────────────────────────────────────────────────
	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// ── Rooms ─────────────────────────────────────────────────────────────────
	mux.Handle("GET /rooms/list", chain(
		http.HandlerFunc(h.Room.List),
		auth, anyRole,
	))
	mux.Handle("POST /rooms/create", chain(
		http.HandlerFunc(h.Room.Create),
		auth, adminOnly,
	))

	// ── Schedules ─────────────────────────────────────────────────────────────
	mux.Handle("POST /rooms/{roomId}/schedule/create", chain(
		http.HandlerFunc(h.Schedule.Create),
		auth, adminOnly,
	))

	// ── Slots ─────────────────────────────────────────────────────────────────
	mux.Handle("GET /rooms/{roomId}/slots/list", chain(
		http.HandlerFunc(h.Slot.ListAvailable),
		auth, anyRole,
	))

	// ── Bookings ──────────────────────────────────────────────────────────────
	mux.Handle("POST /bookings/create", chain(
		http.HandlerFunc(h.Booking.Create),
		auth, userOnly,
	))
	mux.Handle("GET /bookings/list", chain(
		http.HandlerFunc(h.Booking.ListAll),
		auth, adminOnly,
	))
	mux.Handle("GET /bookings/my", chain(
		http.HandlerFunc(h.Booking.ListMy),
		auth, userOnly,
	))
	mux.Handle("POST /bookings/{bookingId}/cancel", chain(
		http.HandlerFunc(h.Booking.Cancel),
		auth, userOnly,
	))

	return mux
}

// info godoc
// @Summary     Проверка доступности сервиса
// @Description Возвращает 200 OK, если сервис запущен и работает.
// @Tags        system
// @Produce     json
// @Success     200 {object} map[string]string
// @Router      /_info [get]
func info(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
