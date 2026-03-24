package handler

import (
	"net/http"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/ports"
)

type AuthHandler struct {
	authSvc ports.AuthService
}

func NewAuthHandler(authSvc ports.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// DummyLogin godoc
// @Summary     Получить тестовый JWT по роли
// @Description Выдаёт подписанный JWT для указанной роли с фиксированным user_id.
//
//	Для роли admin всегда возвращается ID 00000000-0000-0000-0000-000000000001;
//	для роли user — 00000000-0000-0000-0000-000000000002.
//
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body dummyLoginRequest true "Роль пользователя"
// @Success     200 {object} tokenResponse
// @Failure     400 {object} errorResponse
// @Router      /dummyLogin [post]
func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req dummyLoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}
	if err := req.validate(); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	token, err := h.authSvc.DummyLogin(r.Context(), entity.Role(req.Role))
	if err != nil {
		writeInternalError(w)
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{Token: token})
}

// Register godoc
// @Summary     Регистрация нового пользователя
// @Description Создаёт учётную запись пользователя с хешированным паролем (bcrypt) и возвращает данные созданного пользователя.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body registerRequest true "Данные для регистрации"
// @Success     201 {object} userResponse
// @Failure     400 {object} errorResponse "Ошибка валидации или email уже занят"
// @Router      /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}
	if err := req.validate(); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	user, err := h.authSvc.Register(r.Context(), req.Email, req.Password, entity.Role(req.Role))
	if err != nil {
		if !writeDomainError(w, err) {
			writeInternalError(w)
		}
		return
	}

	writeJSON(w, http.StatusCreated, userResponse{User: userToDTO(*user)})
}

// Login godoc
// @Summary     Авторизация по email и паролю
// @Description Проверяет учётные данные и возвращает подписанный JWT при успешной аутентификации.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body loginRequest true "Учётные данные"
// @Success     200 {object} tokenResponse
// @Failure     400 {object} errorResponse "Ошибка валидации"
// @Failure     401 {object} errorResponse "Неверные учётные данные"
// @Router      /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}
	if err := req.validate(); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	token, err := h.authSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if !writeDomainError(w, err) {
			writeInternalError(w)
		}
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{Token: token})
}
