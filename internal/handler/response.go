package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
)

func decodeJSON(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	return nil
}

func queryInt(r *http.Request, key string, def int) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return def, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parameter %q must be an integer", key)
	}
	return v, nil
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Error: errorBody{Code: code, Message: message}})
}

func writeInternalError(w http.ResponseWriter) {
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}

func writeDomainError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, entity.ErrRoomNotFound):
		writeError(w, http.StatusNotFound, "ROOM_NOT_FOUND", err.Error())
	case errors.Is(err, entity.ErrScheduleExists):
		writeError(w, http.StatusConflict, "SCHEDULE_EXISTS", err.Error())
	case errors.Is(err, entity.ErrSlotNotFound):
		writeError(w, http.StatusNotFound, "SLOT_NOT_FOUND", err.Error())
	case errors.Is(err, entity.ErrSlotAlreadyBooked):
		writeError(w, http.StatusConflict, "SLOT_ALREADY_BOOKED", err.Error())
	case errors.Is(err, entity.ErrSlotInPast):
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	case errors.Is(err, entity.ErrBookingNotFound):
		writeError(w, http.StatusNotFound, "BOOKING_NOT_FOUND", err.Error())
	case errors.Is(err, entity.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
	case errors.Is(err, entity.ErrEmailAlreadyTaken):
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	case errors.Is(err, entity.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	default:
		return false
	}
	return true
}
