package handler

import (
	"RIP-WEB/internal/app/ds"
	"RIP-WEB/internal/app/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest представляет запрос на аутентификацию
type LoginRequest struct {
	Login    string `json:"login" binding:"required" example:"research_user"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginResponse представляет ответ на аутентификацию
type LoginResponse struct {
	Message string       `json:"message" example:"Успешная аутентификация"`
	Token   string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User    UserResponse `json:"user"`
}

// RegisterRequest представляет запрос на регистрацию
type RegisterRequest struct {
	Login    string `json:"login" binding:"required,min=3,max=25" example:"new_user"`
	Password string `json:"password" binding:"required,min=6" example:"securepassword"`
}

// UserResponse представляет данные пользователя
type UserResponse struct {
	ID          uint   `json:"id" example:"1"`
	Login       string `json:"login" example:"research_user"`
	IsModerator bool   `json:"is_moderator" example:"false"`
}

// ErrorResponse представляет ошибку
type ErrorResponse struct {
	Error string `json:"error" example:"Описание ошибки"`
}

// MessageResponse представляет сообщение
type MessageResponse struct {
	Message string `json:"message" example:"Сообщение об успехе"`
}

// RegisterUser godoc
// @Summary Регистрация нового пользователя
// @Description Создание учетной записи пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "Данные для регистрации"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/users/register [post]
func (h *Handler) RegisterUser(ctx *gin.Context) {
	var request RegisterRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Проверяем существует ли пользователь
	existingUser, err := h.Repository.GetUserByLogin(request.Login)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Ошибка сервера"})
		return
	}
	if existingUser != nil {
		ctx.JSON(http.StatusConflict, ErrorResponse{Error: "Пользователь с таким логином уже существует"})
		return
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Ошибка хэширования пароля"})
		return
	}

	// Создаем нового пользователя
	user := ds.Users{
		Login:       request.Login,
		Password:    string(hashedPassword),
		IsModerator: false,
	}

	if err := h.Repository.CreateUser(&user); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, UserResponse{
		ID:          user.ID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
	})
}

// LoginUser godoc
// @Summary Аутентификация пользователя
// @Description Вход в систему с получением JWT токена
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Данные для входа"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/users/login [post]
func (h *Handler) LoginUser(ctx *gin.Context) {
	var request LoginRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.Repository.GetUserByLogin(request.Login)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Ошибка сервера"})
		return
	}
	if user == nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Неверный логин или пароль"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Неверный логин или пароль"})
		return
	}

	// Генерируем JWT
	token, err := utils.GenerateJWT(user, h.Config.JWTSecret, h.Config.JWTExpiration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Ошибка генерации токена"})
		return
	}

	// Устанавливаем куки
	ctx.SetCookie("token", token, h.Config.JWTExpiration*3600, "/", "", false, true)

	ctx.JSON(http.StatusOK, LoginResponse{
		Message: "Успешная аутентификация",
		Token:   token,
		User: UserResponse{
			ID:          user.ID,
			Login:       user.Login,
			IsModerator: user.IsModerator,
		},
	})
}

// LogoutUser godoc
// @Summary Выход из системы
// @Description Завершение сессии пользователя с добавлением токена в blacklist
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} MessageResponse
// @Router /api/users/logout [post]
func (h *Handler) LogoutUser(ctx *gin.Context) {
	token, exists := ctx.Get("token")
	if exists && token != "" {
		// Добавляем токен в blacklist на оставшееся время
		claims, err := utils.ValidateJWT(token.(string), h.Config.JWTSecret)
		if err == nil {
			expiration := time.Until(claims.ExpiresAt.Time)
			if expiration > 0 {
				h.TokenManager.AddToBlacklist(token.(string), expiration)
			}
		}
	}

	// Удаляем куки
	ctx.SetCookie("token", "", -1, "/", "", false, true)

	ctx.JSON(http.StatusOK, MessageResponse{
		Message: "Успешный выход из системы",
	})
}

// GetCurrentUser godoc
// @Summary Получение информации о текущем пользователе
// @Description Возвращает данные авторизованного пользователя
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/users/me [get]
func (h *Handler) GetCurrentUser(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Требуется аутентификация"})
		return
	}

	user, err := h.Repository.GetUserByID(userID.(uint))
	if err != nil || user == nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Пользователь не найден"})
		return
	}

	ctx.JSON(http.StatusOK, UserResponse{
		ID:          user.ID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
	})
}

// UpdateUserProfile godoc
// @Summary Обновление профиля пользователя
// @Description Обновляет данные текущего пользователя
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body UpdateUserRequest true "Данные для обновления"
// @Success 200 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/users/profile [put]
func (h *Handler) UpdateUserProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Требуется аутентификация"})
		return
	}

	var request struct {
		Login string `json:"login"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.Repository.GetUserByID(userID.(uint))
	if err != nil || user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	if request.Login != "" {
		// Проверяем не занят ли логин другим пользователем
		var existingUser ds.Users
		err := h.Repository.GetDB().Where("login = ? AND id != ?", request.Login, user.ID).First(&existingUser).Error
		if err == nil {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Логин уже занят"})
			return
		}
		user.Login = request.Login
	}

	if err := h.Repository.GetDB().Save(user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, UserResponse{
		ID:          user.ID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
	})
}

// UpdateUserRequest представляет запрос на обновление пользователя
type UpdateUserRequest struct {
	Login string `json:"login" example:"new_login"`
}
