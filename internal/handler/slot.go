package handler

import (
	"net/http"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/ports"
	"github.com/google/uuid"
)

type SlotHandler struct {
	slotSvc ports.SlotService
	roomSvc ports.RoomService
}

func NewSlotHandler(slotSvc ports.SlotService, roomSvc ports.RoomService) *SlotHandler {
	return &SlotHandler{slotSvc: slotSvc, roomSvc: roomSvc}
}

// ListAvailable godoc
// @Summary     Список доступных слотов для бронирования
// @Description Возвращает все незанятые 30-минутные слоты переговорки на указанную дату (UTC).
//
//	Наиболее нагруженный эндпоинт; время ответа оптимизировано до < 200 мс.
//
// @Tags        slots
// @Produce     json
// @Security    BearerAuth
// @Param       roomId path  string true "UUID переговорки"
// @Param       date   query string true "Дата в формате YYYY-MM-DD (UTC)"
// @Success     200    {object} slotListResponse
// @Failure     400    {object} errorResponse "Отсутствует или некорректен параметр date"
// @Failure     401    {object} errorResponse "Отсутствует или недействителен токен"
// @Failure     403    {object} errorResponse "Недостаточно прав"
// @Failure     404    {object} errorResponse "Переговорка не найдена"
// @Router      /rooms/{roomId}/slots/list [get]
func (h *SlotHandler) ListAvailable(w http.ResponseWriter, r *http.Request) {
	roomID, err := uuid.Parse(r.PathValue("roomId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid roomId")
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "date query parameter is required")
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "date must be in YYYY-MM-DD format")
		return
	}

	if _, err := h.roomSvc.GetByID(r.Context(), roomID); err != nil {
		writeError(w, http.StatusNotFound, "ROOM_NOT_FOUND", "room not found")
		return
	}

	slots, err := h.slotSvc.ListAvailable(r.Context(), roomID, date)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	dtos := make([]slotDTO, len(slots))
	for i, slot := range slots {
		dtos[i] = slotToDTO(slot)
	}
	writeJSON(w, http.StatusOK, slotListResponse{Slots: dtos})
}
