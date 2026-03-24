package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/handler"
	jwtpkg "github.com/avito-internships/test-backend-1-cQu1x/internal/infrastructure/jwt"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/testutil"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/usecase"
	bookingSvc "github.com/avito-internships/test-backend-1-cQu1x/internal/usecase/booking"
	roomSvc "github.com/avito-internships/test-backend-1-cQu1x/internal/usecase/room"
	scheduleSvc "github.com/avito-internships/test-backend-1-cQu1x/internal/usecase/schedule"
	slotSvc "github.com/avito-internships/test-backend-1-cQu1x/internal/usecase/slot"
)

func newTestServer(t *testing.T) (*httptest.Server, *jwtpkg.TokenManager) {
	t.Helper()

	store := testutil.NewInMemStore()
	tm := jwtpkg.NewTokenManager("test-secret")

	authUC := usecase.NewAuthUseCase(store, tm)
	roomService := roomSvc.NewService(store)
	slotService := slotSvc.NewService(store, store)
	scheduleService := scheduleSvc.NewService(store, store, slotService)
	bookingService := bookingSvc.NewService(store, store)

	h := handler.Handlers{
		Auth:     handler.NewAuthHandler(authUC),
		Room:     handler.NewRoomHandler(roomService),
		Schedule: handler.NewScheduleHandler(scheduleService),
		Slot:     handler.NewSlotHandler(slotService, roomService),
		Booking:  handler.NewBookingHandler(bookingService),
	}

	return httptest.NewServer(handler.NewRouter(h, tm)), tm
}

// ── HTTP helpers ──────────────────────────────────────────────────────────────

func post(t *testing.T, srv *httptest.Server, path, token string, body any) (int, map[string]any) {
	t.Helper()
	return doRequest(t, srv, http.MethodPost, path, token, body)
}

func get(t *testing.T, srv *httptest.Server, path, token string) (int, map[string]any) {
	t.Helper()
	return doRequest(t, srv, http.MethodGet, path, token, nil)
}

func doRequest(t *testing.T, srv *httptest.Server, method, path, token string, body any) (int, map[string]any) {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode request body: %v", err)
		}
	}

	req, err := http.NewRequest(method, srv.URL+path, &buf)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp.StatusCode, result
}

// ── Step helpers ──────────────────────────────────────────────────────────────

func stepDummyToken(t *testing.T, srv *httptest.Server, role string) string {
	t.Helper()
	status, body := post(t, srv, "/dummyLogin", "", map[string]string{"role": role})
	if status != http.StatusOK {
		t.Fatalf("dummyLogin failed: status=%d body=%v", status, body)
	}
	token, ok := body["token"].(string)
	if !ok || token == "" {
		t.Fatalf("expected token in response, got: %v", body)
	}
	return token
}

func stepCreateRoom(t *testing.T, srv *httptest.Server, token, name string) string {
	t.Helper()
	status, body := post(t, srv, "/rooms/create", token, map[string]any{"name": name})
	if status != http.StatusCreated {
		t.Fatalf("create room: expected 201, got %d: %v", status, body)
	}
	roomID, _ := body["room"].(map[string]any)["id"].(string)
	if roomID == "" {
		t.Fatalf("expected room ID in response, got: %v", body)
	}
	return roomID
}

func stepCreateSchedule(t *testing.T, srv *httptest.Server, token, roomID string, daysOfWeek []int, startTime, endTime string) {
	t.Helper()
	status, body := post(t, srv, fmt.Sprintf("/rooms/%s/schedule/create", roomID), token, map[string]any{
		"daysOfWeek": daysOfWeek,
		"startTime":  startTime,
		"endTime":    endTime,
	})
	if status != http.StatusCreated {
		t.Fatalf("create schedule: expected 201, got %d: %v", status, body)
	}
}

func stepListSlots(t *testing.T, srv *httptest.Server, token, roomID, date string) []any {
	t.Helper()
	status, body := get(t, srv, fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomID, date), token)
	if status != http.StatusOK {
		t.Fatalf("list slots: expected 200, got %d: %v", status, body)
	}
	slots, _ := body["slots"].([]any)
	return slots
}

func stepCreateBooking(t *testing.T, srv *httptest.Server, token, slotID string) (string, map[string]any) {
	t.Helper()
	status, body := post(t, srv, "/bookings/create", token, map[string]any{"slotId": slotID})
	if status != http.StatusCreated {
		t.Fatalf("create booking: expected 201, got %d: %v", status, body)
	}
	bookingObj, _ := body["booking"].(map[string]any)
	bookingID, _ := bookingObj["id"].(string)
	if bookingID == "" {
		t.Fatalf("expected booking ID in response, got: %v", body)
	}
	return bookingID, bookingObj
}

func stepCancelBooking(t *testing.T, srv *httptest.Server, token, bookingID string) map[string]any {
	t.Helper()
	status, body := post(t, srv, fmt.Sprintf("/bookings/%s/cancel", bookingID), token, map[string]any{})
	if status != http.StatusOK {
		t.Fatalf("cancel booking: expected 200, got %d: %v", status, body)
	}
	return body["booking"].(map[string]any)
}

func stepAssertSlotAbsent(t *testing.T, slots []any, slotID string) {
	t.Helper()
	for _, s := range slots {
		if s.(map[string]any)["id"] == slotID {
			t.Error("booked slot must not appear in available slots list")
		}
	}
}

func stepAssertSlotPresent(t *testing.T, slots []any, slotID string) {
	t.Helper()
	for _, s := range slots {
		if s.(map[string]any)["id"] == slotID {
			return
		}
	}
	t.Error("expected slot to be present in available slots list")
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestIntegration_CreateRoomScheduleBooking(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	adminToken := stepDummyToken(t, srv, "admin")
	userToken := stepDummyToken(t, srv, "user")

	roomID := stepCreateRoom(t, srv, adminToken, "Integration Room")

	stepCreateSchedule(t, srv, adminToken, roomID, []int{1, 2, 3, 4, 5, 6, 7}, "09:00", "09:30")

	tomorrow := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	slots := stepListSlots(t, srv, userToken, roomID, tomorrow)
	if len(slots) == 0 {
		t.Fatal("expected at least one available slot for tomorrow, got none")
	}
	slotID, _ := slots[0].(map[string]any)["id"].(string)

	_, bookingObj := stepCreateBooking(t, srv, userToken, slotID)
	if bookingObj["status"] != "active" {
		t.Errorf("expected booking status 'active', got %v", bookingObj["status"])
	}
	if bookingObj["slotId"] != slotID {
		t.Errorf("expected slotId %v, got %v", slotID, bookingObj["slotId"])
	}

	slotsAfter := stepListSlots(t, srv, userToken, roomID, tomorrow)
	stepAssertSlotAbsent(t, slotsAfter, slotID)
}

func TestIntegration_CancelBooking(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	adminToken := stepDummyToken(t, srv, "admin")
	userToken := stepDummyToken(t, srv, "user")

	roomID := stepCreateRoom(t, srv, adminToken, "Cancel Test Room")
	stepCreateSchedule(t, srv, adminToken, roomID, []int{1, 2, 3, 4, 5, 6, 7}, "10:00", "10:30")

	tomorrow := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	slots := stepListSlots(t, srv, userToken, roomID, tomorrow)
	if len(slots) == 0 {
		t.Fatal("expected available slot, got none")
	}
	slotID, _ := slots[0].(map[string]any)["id"].(string)

	bookingID, _ := stepCreateBooking(t, srv, userToken, slotID)

	booking := stepCancelBooking(t, srv, userToken, bookingID)
	if booking["status"] != "cancelled" {
		t.Errorf("expected status 'cancelled', got %v", booking["status"])
	}

	// Idempotent second cancellation must also return 200.
	booking = stepCancelBooking(t, srv, userToken, bookingID)
	if booking["status"] != "cancelled" {
		t.Errorf("expected status 'cancelled' on second cancel, got %v", booking["status"])
	}

	slotsAfterCancel := stepListSlots(t, srv, userToken, roomID, tomorrow)
	stepAssertSlotPresent(t, slotsAfterCancel, slotID)
}

func TestInfo(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	status, _ := get(t, srv, "/_info", "")
	if status != http.StatusOK {
		t.Errorf("/_info expected 200, got %d", status)
	}
}

func TestDuplicateBooking_SlotAlreadyBooked(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	adminToken := stepDummyToken(t, srv, "admin")
	userToken := stepDummyToken(t, srv, "user")

	roomID := stepCreateRoom(t, srv, adminToken, "Double Book Room")
	stepCreateSchedule(t, srv, adminToken, roomID, []int{1, 2, 3, 4, 5, 6, 7}, "11:00", "11:30")

	tomorrow := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	slots := stepListSlots(t, srv, userToken, roomID, tomorrow)
	slotID, _ := slots[0].(map[string]any)["id"].(string)

	stepCreateBooking(t, srv, userToken, slotID)

	// Second booking on the same slot must fail with 409.
	status, body := post(t, srv, "/bookings/create", userToken, map[string]any{"slotId": slotID})
	if status != http.StatusConflict {
		t.Errorf("duplicate booking expected 409, got %d: %v", status, body)
	}
}

func TestCreateSchedule_DuplicateRejected(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	adminToken := stepDummyToken(t, srv, "admin")

	roomID := stepCreateRoom(t, srv, adminToken, "Immutable Room")
	stepCreateSchedule(t, srv, adminToken, roomID, []int{1}, "09:00", "09:30")

	// Second schedule for the same room must fail with 409.
	status, body := post(t, srv, fmt.Sprintf("/rooms/%s/schedule/create", roomID), adminToken, map[string]any{
		"daysOfWeek": []int{1},
		"startTime":  "09:00",
		"endTime":    "09:30",
	})
	if status != http.StatusConflict {
		t.Errorf("duplicate schedule expected 409, got %d: %v", status, body)
	}
}

func TestAuth_RoleEnforcement(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	userToken := stepDummyToken(t, srv, "user")

	// A user must not be able to create a room.
	status, _ := post(t, srv, "/rooms/create", userToken, map[string]any{"name": "Forbidden Room"})
	if status != http.StatusForbidden {
		t.Errorf("expected 403 for user creating room, got %d", status)
	}

	// A request without a token must be rejected.
	status, _ = get(t, srv, "/rooms/list", "")
	if status != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", status)
	}
}
