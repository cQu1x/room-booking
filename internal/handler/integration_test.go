// Package handler_test contains integration/E2E tests for the full HTTP stack.
// The real service implementations are wired with an in-memory store so no
// database is required.
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

// newTestServer assembles the complete HTTP server backed by an in-memory store.
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

// post sends a POST request with a JSON body and returns the parsed response.
func post(t *testing.T, srv *httptest.Server, path, token string, body any) (int, map[string]any) {
	t.Helper()
	return doRequest(t, srv, http.MethodPost, path, token, body)
}

// get sends a GET request and returns the parsed response.
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

// dummyToken fetches a token for the given role via /dummyLogin.
func dummyToken(t *testing.T, srv *httptest.Server, role string) string {
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

// TestIntegration_CreateRoomScheduleBooking covers the full happy path:
// admin creates room → admin creates schedule → user books a slot.
func TestIntegration_CreateRoomScheduleBooking(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	adminToken := dummyToken(t, srv, "admin")
	userToken := dummyToken(t, srv, "user")

	// Step 1: Admin creates a room.
	status, body := post(t, srv, "/rooms/create", adminToken, map[string]any{
		"name": "Integration Room",
	})
	if status != http.StatusCreated {
		t.Fatalf("create room: expected 201, got %d: %v", status, body)
	}
	roomObj, _ := body["room"].(map[string]any)
	roomID, _ := roomObj["id"].(string)
	if roomID == "" {
		t.Fatalf("expected room ID in response, got: %v", body)
	}

	// Step 2: Admin creates a schedule for the room (all 7 days, 09:00–09:30).
	// Use all days so the slot window always covers today.
	status, body = post(t, srv, fmt.Sprintf("/rooms/%s/schedule/create", roomID), adminToken, map[string]any{
		"daysOfWeek": []int{1, 2, 3, 4, 5, 6, 7},
		"startTime":  "09:00",
		"endTime":    "09:30",
	})
	if status != http.StatusCreated {
		t.Fatalf("create schedule: expected 201, got %d: %v", status, body)
	}

	// Step 3: User lists available slots for tomorrow (always future).
	tomorrow := time.Now().UTC().AddDate(0, 0, 1)
	dateStr := tomorrow.Format("2006-01-02")
	status, body = get(t, srv, fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomID, dateStr), userToken)
	if status != http.StatusOK {
		t.Fatalf("list slots: expected 200, got %d: %v", status, body)
	}
	slotsRaw, _ := body["slots"].([]any)
	if len(slotsRaw) == 0 {
		t.Fatalf("expected at least one available slot for tomorrow, got none")
	}
	firstSlot, _ := slotsRaw[0].(map[string]any)
	slotID, _ := firstSlot["id"].(string)
	if slotID == "" {
		t.Fatalf("expected slot ID, got: %v", firstSlot)
	}

	// Step 4: User creates a booking for the slot.
	status, body = post(t, srv, "/bookings/create", userToken, map[string]any{
		"slotId": slotID,
	})
	if status != http.StatusCreated {
		t.Fatalf("create booking: expected 201, got %d: %v", status, body)
	}
	bookingObj, _ := body["booking"].(map[string]any)
	bookingID, _ := bookingObj["id"].(string)
	if bookingID == "" {
		t.Fatalf("expected booking ID in response, got: %v", body)
	}
	if bookingObj["status"] != "active" {
		t.Errorf("expected booking status 'active', got %v", bookingObj["status"])
	}
	if bookingObj["slotId"] != slotID {
		t.Errorf("expected slotId %v, got %v", slotID, bookingObj["slotId"])
	}

	// Step 5: Confirm the slot is no longer available.
	status, body = get(t, srv, fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomID, dateStr), userToken)
	if status != http.StatusOK {
		t.Fatalf("list slots after booking: expected 200, got %d", status)
	}
	slotsAfter, _ := body["slots"].([]any)
	for _, s := range slotsAfter {
		sm, _ := s.(map[string]any)
		if sm["id"] == slotID {
			t.Error("booked slot must not appear in available slots list")
		}
	}
}

// TestIntegration_CancelBooking verifies that a user can cancel a booking
// and that repeating the cancellation is idempotent.
func TestIntegration_CancelBooking(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	adminToken := dummyToken(t, srv, "admin")
	userToken := dummyToken(t, srv, "user")

	// Setup: create room and schedule.
	status, body := post(t, srv, "/rooms/create", adminToken, map[string]any{
		"name": "Cancel Test Room",
	})
	if status != http.StatusCreated {
		t.Fatalf("create room: %d %v", status, body)
	}
	roomID := body["room"].(map[string]any)["id"].(string)

	status, body = post(t, srv, fmt.Sprintf("/rooms/%s/schedule/create", roomID), adminToken, map[string]any{
		"daysOfWeek": []int{1, 2, 3, 4, 5, 6, 7},
		"startTime":  "10:00",
		"endTime":    "10:30",
	})
	if status != http.StatusCreated {
		t.Fatalf("create schedule: %d %v", status, body)
	}

	// Get a future slot and book it.
	tomorrow := time.Now().UTC().AddDate(0, 0, 1)
	dateStr := tomorrow.Format("2006-01-02")
	status, body = get(t, srv, fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomID, dateStr), userToken)
	if status != http.StatusOK || len(body["slots"].([]any)) == 0 {
		t.Fatalf("expected available slot, got %d %v", status, body)
	}
	slotID := body["slots"].([]any)[0].(map[string]any)["id"].(string)

	status, body = post(t, srv, "/bookings/create", userToken, map[string]any{
		"slotId": slotID,
	})
	if status != http.StatusCreated {
		t.Fatalf("create booking: %d %v", status, body)
	}
	bookingID := body["booking"].(map[string]any)["id"].(string)

	// Cancel the booking.
	status, body = post(t, srv, fmt.Sprintf("/bookings/%s/cancel", bookingID), userToken, map[string]any{})
	if status != http.StatusOK {
		t.Fatalf("cancel booking: expected 200, got %d: %v", status, body)
	}
	if body["booking"].(map[string]any)["status"] != "cancelled" {
		t.Errorf("expected status 'cancelled', got %v", body["booking"].(map[string]any)["status"])
	}

	// Idempotent second cancellation must also return 200.
	status, body = post(t, srv, fmt.Sprintf("/bookings/%s/cancel", bookingID), userToken, map[string]any{})
	if status != http.StatusOK {
		t.Fatalf("idempotent cancel: expected 200, got %d: %v", status, body)
	}
	if body["booking"].(map[string]any)["status"] != "cancelled" {
		t.Errorf("expected status 'cancelled' on second cancel, got %v", body["booking"].(map[string]any)["status"])
	}

	// After cancellation the slot should be available again.
	status, body = get(t, srv, fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomID, dateStr), userToken)
	if status != http.StatusOK {
		t.Fatalf("list slots after cancel: %d", status)
	}
	found := false
	for _, s := range body["slots"].([]any) {
		if s.(map[string]any)["id"] == slotID {
			found = true
			break
		}
	}
	if !found {
		t.Error("cancelled slot should be available again")
	}
}

// TestInfo verifies the /_info health-check endpoint.
func TestInfo(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	status, _ := get(t, srv, "/_info", "")
	if status != http.StatusOK {
		t.Errorf("/_info expected 200, got %d", status)
	}
}

// TestDuplicateBooking_SlotAlreadyBooked ensures a second booking on the same slot is rejected.
func TestDuplicateBooking_SlotAlreadyBooked(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	adminToken := dummyToken(t, srv, "admin")
	userToken := dummyToken(t, srv, "user")

	status, body := post(t, srv, "/rooms/create", adminToken, map[string]any{"name": "Double Book Room"})
	if status != http.StatusCreated {
		t.Fatalf("create room: %d %v", status, body)
	}
	roomID := body["room"].(map[string]any)["id"].(string)

	post(t, srv, fmt.Sprintf("/rooms/%s/schedule/create", roomID), adminToken, map[string]any{
		"daysOfWeek": []int{1, 2, 3, 4, 5, 6, 7},
		"startTime":  "11:00",
		"endTime":    "11:30",
	})

	tomorrow := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	_, slotsBody := get(t, srv, fmt.Sprintf("/rooms/%s/slots/list?date=%s", roomID, tomorrow), userToken)
	slotID := slotsBody["slots"].([]any)[0].(map[string]any)["id"].(string)

	// First booking succeeds.
	status, _ = post(t, srv, "/bookings/create", userToken, map[string]any{"slotId": slotID})
	if status != http.StatusCreated {
		t.Fatalf("first booking expected 201, got %d", status)
	}

	// Second booking on the same slot must fail with 409.
	status, body = post(t, srv, "/bookings/create", userToken, map[string]any{"slotId": slotID})
	if status != http.StatusConflict {
		t.Errorf("duplicate booking expected 409, got %d: %v", status, body)
	}
}

// TestCreateSchedule_DuplicateRejected verifies the immutability constraint.
func TestCreateSchedule_DuplicateRejected(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	adminToken := dummyToken(t, srv, "admin")

	status, body := post(t, srv, "/rooms/create", adminToken, map[string]any{"name": "Immutable Room"})
	if status != http.StatusCreated {
		t.Fatalf("create room: %d %v", status, body)
	}
	roomID := body["room"].(map[string]any)["id"].(string)

	schedulePayload := map[string]any{
		"daysOfWeek": []int{1},
		"startTime":  "09:00",
		"endTime":    "09:30",
	}

	status, _ = post(t, srv, fmt.Sprintf("/rooms/%s/schedule/create", roomID), adminToken, schedulePayload)
	if status != http.StatusCreated {
		t.Fatalf("first schedule expected 201, got %d", status)
	}

	status, body = post(t, srv, fmt.Sprintf("/rooms/%s/schedule/create", roomID), adminToken, schedulePayload)
	if status != http.StatusConflict {
		t.Errorf("duplicate schedule expected 409, got %d: %v", status, body)
	}
}

// TestAuth_RoleEnforcement verifies that role-based access control works.
func TestAuth_RoleEnforcement(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	userToken := dummyToken(t, srv, "user")

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
