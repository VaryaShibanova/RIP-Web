package handler

import (
	"RIP-WEB/internal/app/ds"
	"RIP-WEB/internal/app/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

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
	var request struct {
		Login    string `json:"login" binding:"required,min=3,max=25"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем существует ли пользователь
	existingUser, err := h.Repository.GetUserByLogin(request.Login)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}
	if existingUser != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Пользователь с таким логином уже существует"})
		return
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка хэширования пароля"})
		return
	}

	// Создаем нового пользователя
	user := ds.Users{
		Login:       request.Login,
		Password:    string(hashedPassword),
		IsModerator: false,
	}

	if err := h.Repository.CreateUser(&user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Не возвращаем пароль
	user.Password = ""

	ctx.JSON(http.StatusCreated, user)
}

// GetUserProfile godoc
// @Summary Получение информации о текущем пользователе
// @Description Возвращает данные авторизованного пользователя
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/users/me [get]
func (h *Handler) GetUserProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		return
	}

	user, err := h.Repository.GetUserByID(userID.(uint))
	if err != nil || user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	// Не возвращаем пароль
	user.Password = ""

	ctx.JSON(http.StatusOK, user)
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
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
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

	// Не возвращаем пароль
	user.Password = ""

	ctx.JSON(http.StatusOK, user)
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
	var request struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.Repository.GetUserByLogin(request.Login)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}
	if user == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
		return
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
		return
	}

	// Генерируем JWT токен
	token, err := h.generateJWT(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
		return
	}

	// Устанавливаем куки
	ctx.SetCookie("token", token, h.Config.JWTExpiration*3600, "/", "", false, true)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Успешная аутентификация",
		"token":   token,
		"user": gin.H{
			"id":           user.ID,
			"login":        user.Login,
			"is_moderator": user.IsModerator,
		},
	})
}

// LogoutUser godoc
// @Summary Выход из системы
// @Description Завершение сессии пользователя
// @Tags auth
// @Produce json
// @Success 200 {object} MessageResponse
// @Router /api/users/logout [post]
func (h *Handler) LogoutUser(ctx *gin.Context) {
	// Удаляем куки
	ctx.SetCookie("token", "", -1, "/", "", false, true)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Успешный выход из системы",
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
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		return
	}

	user, err := h.Repository.GetUserByID(userID.(uint))
	if err != nil || user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	// Не возвращаем пароль
	user.Password = ""

	ctx.JSON(http.StatusOK, user)
}

// Вспомогательная функция для генерации JWT
func (h *Handler) generateJWT(user *ds.Users) (string, error) {
	// Используем утилиту из utils/jwt.go
	return utils.GenerateJWT(user, h.Config.JWTSecret, h.Config.JWTExpiration)
}
