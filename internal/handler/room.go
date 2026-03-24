package handler

import (
	"net/http"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/ports"
)

type RoomHandler struct {
	roomSvc ports.RoomService
}

func NewRoomHandler(roomSvc ports.RoomService) *RoomHandler {
	return &RoomHandler{roomSvc: roomSvc}
}

// List godoc
// @Summary     Список переговорок
// @Description Возвращает список всех переговорок в системе.
// @Tags        rooms
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} roomListResponse
// @Failure     401 {object} errorResponse "Отсутствует или недействителен токен"
// @Failure     403 {object} errorResponse "Недостаточно прав"
// @Router      /rooms/list [get]
func (h *RoomHandler) List(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.roomSvc.List(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}

	dtos := make([]roomDTO, len(rooms))
	for i, room := range rooms {
		dtos[i] = roomToDTO(room)
	}
	writeJSON(w, http.StatusOK, roomListResponse{Rooms: dtos})
}

// Create godoc
// @Summary     Создать переговорку
// @Description Создаёт новую переговорку. Требуется роль admin.
// @Tags        rooms
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body createRoomRequest true "Данные переговорки"
// @Success     201 {object} roomResponse
// @Failure     400 {object} errorResponse "Ошибка валидации"
// @Failure     401 {object} errorResponse "Отсутствует или недействителен токен"
// @Failure     403 {object} errorResponse "Требуется роль admin"
// @Router      /rooms/create [post]
func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRoomRequest
	if err := decodeJSON(r, w, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}
	if err := req.validate(); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	room, err := h.roomSvc.Create(r.Context(), req.Name, req.Description, req.Capacity)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, roomResponse{Room: roomToDTO(*room)})
}
