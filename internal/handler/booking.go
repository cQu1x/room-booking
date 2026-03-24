package handler

import (
	"errors"
	"net/http"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/ports"
	"github.com/google/uuid"
)

const (
	minPageSize = 1
	maxPageSize = 100
)

type BookingHandler struct {
	bookingSvc ports.BookingService
}

func NewBookingHandler(bookingSvc ports.BookingService) *BookingHandler {
	return &BookingHandler{bookingSvc: bookingSvc}
}

// Create godoc
// @Summary     Создать бронь на слот
// @Description Создаёт активную бронь на указанный слот от имени аутентифицированного пользователя.
//
//	Слот должен существовать, не находиться в прошлом и не иметь активной брони.
//	Опционально генерирует мок-ссылку на конференцию.
//
// @Tags        bookings
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body createBookingRequest true "Данные брони"
// @Success     201  {object} bookingResponse
// @Failure     400  {object} errorResponse "Ошибка валидации или слот в прошлом"
// @Failure     401  {object} errorResponse "Отсутствует или недействителен токен"
// @Failure     403  {object} errorResponse "Требуется роль user"
// @Failure     404  {object} errorResponse "Слот не найден"
// @Failure     409  {object} errorResponse "Слот уже забронирован"
// @Router      /bookings/create [post]
func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createBookingRequest
	if err := decodeJSON(r, w, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}
	if err := req.validate(); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	userID := ctxGetUserID(r)
	booking, err := h.bookingSvc.Create(r.Context(), userID, req.SlotID, req.CreateConferenceLink)
	if err != nil {
		if !writeDomainError(w, err) {
			writeInternalError(w, err)
		}
		return
	}

	writeJSON(w, http.StatusCreated, bookingResponse{Booking: bookingToDTO(*booking)})
}

// ListAll godoc
// @Summary     Список всех броней (только admin)
// @Description Возвращает постраничный список всех броней в системе. Требуется роль admin.
// @Tags        bookings
// @Produce     json
// @Security    BearerAuth
// @Param       page     query int false "Номер страницы (по умолчанию: 1)"
// @Param       pageSize query int false "Размер страницы, 1–100 (по умолчанию: 20)"
// @Success     200 {object} bookingListResponse
// @Failure     400 {object} errorResponse "Некорректные параметры пагинации"
// @Failure     401 {object} errorResponse "Отсутствует или недействителен токен"
// @Failure     403 {object} errorResponse "Требуется роль admin"
// @Router      /bookings/list [get]
func (h *BookingHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	page, err := queryInt(r, "page", 1)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	pageSize, err := queryInt(r, "pageSize", 20)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if err := pageSizeInBounds(pageSize); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	bookings, total, err := h.bookingSvc.ListAll(r.Context(), page, pageSize)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	dtos := make([]bookingDTO, len(bookings))
	for i, booking := range bookings {
		dtos[i] = bookingToDTO(booking)
	}
	writeJSON(w, http.StatusOK, bookingListResponse{
		Bookings: dtos,
		Pagination: &paginationDTO{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// ListMy godoc
// @Summary     Мои брони
// @Description Возвращает будущие брони аутентифицированного пользователя (прошедшие слоты не включаются).
// @Tags        bookings
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} bookingListResponse
// @Failure     401 {object} errorResponse "Отсутствует или недействителен токен"
// @Failure     403 {object} errorResponse "Требуется роль user"
// @Router      /bookings/my [get]
func (h *BookingHandler) ListMy(w http.ResponseWriter, r *http.Request) {
	userID := ctxGetUserID(r)
	bookings, err := h.bookingSvc.ListMy(r.Context(), userID)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	dtos := make([]bookingDTO, len(bookings))
	for i, booking := range bookings {
		dtos[i] = bookingToDTO(booking)
	}
	writeJSON(w, http.StatusOK, bookingListResponse{Bookings: dtos})
}

// Cancel godoc
// @Summary     Отменить бронь
// @Description Отменяет бронь. Операция идемпотентна: повторная отмена уже отменённой брони
//
//	возвращает 200 с текущим состоянием, а не ошибку.
//	Отменить можно только свою бронь.
//
// @Tags        bookings
// @Produce     json
// @Security    BearerAuth
// @Param       bookingId path string true "UUID брони"
// @Success     200 {object} bookingResponse
// @Failure     400 {object} errorResponse "Некорректный bookingId"
// @Failure     401 {object} errorResponse "Отсутствует или недействителен токен"
// @Failure     403 {object} errorResponse "Требуется роль user или чужая бронь"
// @Failure     404 {object} errorResponse "Бронь не найдена"
// @Router      /bookings/{bookingId}/cancel [post]
func (h *BookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(r.PathValue("bookingId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid bookingId")
		return
	}

	userID := ctxGetUserID(r)
	booking, err := h.bookingSvc.Cancel(r.Context(), userID, bookingID)
	if err != nil {
		if !writeDomainError(w, err) {
			writeInternalError(w, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, bookingResponse{Booking: bookingToDTO(*booking)})
}

func pageSizeInBounds(pageSize int) error {
	if pageSize < minPageSize || pageSize > maxPageSize {
		return errors.New("pageSize must be between 1 and 100")
	}
	return nil
}
