package handler

import (
	"net/http"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/ports"
	"github.com/google/uuid"
)

type ScheduleHandler struct {
	scheduleSvc ports.ScheduleService
}

func NewScheduleHandler(scheduleSvc ports.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{scheduleSvc: scheduleSvc}
}

// Create godoc
// @Summary     Создать расписание для переговорки
// @Description Задаёт еженедельное расписание доступности переговорки и предварительно генерирует
//
//	30-минутные слоты на ближайшие 7 дней. Расписание можно создать только один раз; изменение невозможно.
//
// @Tags        schedules
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       roomId path     string                true "UUID переговорки"
// @Param       body   body     createScheduleRequest true "Данные расписания"
// @Success     201    {object} scheduleResponse
// @Failure     400    {object} errorResponse "Ошибка валидации"
// @Failure     401    {object} errorResponse "Отсутствует или недействителен токен"
// @Failure     403    {object} errorResponse "Требуется роль admin"
// @Failure     404    {object} errorResponse "Переговорка не найдена"
// @Failure     409    {object} errorResponse "Расписание для переговорки уже существует"
// @Router      /rooms/{roomId}/schedule/create [post]
func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	roomID, err := uuid.Parse(r.PathValue("roomId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid roomId")
		return
	}

	var req createScheduleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}
	if err := req.validate(); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	schedule, err := h.scheduleSvc.Create(r.Context(), roomID, req.DaysOfWeek, req.StartTime, req.EndTime)
	if err != nil {
		if !writeDomainError(w, err) {
			writeInternalError(w)
		}
		return
	}

	writeJSON(w, http.StatusCreated, scheduleResponse{Schedule: scheduleToDTO(*schedule)})
}
