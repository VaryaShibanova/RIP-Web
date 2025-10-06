package handler

import (
	"RIP-WEB/internal/app/ds"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ДОМЕН ПОЛЬЗОВАТЕЛЯ (USERS)

// RegisterUser - POST регистрация
func (h *Handler) RegisterUser(ctx *gin.Context) {
	var request struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем существует ли пользователь
	var existingUser ds.Users
	err := h.Repository.GetDB().Where("login = ?", request.Login).First(&existingUser).Error
	if err == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Пользователь с таким логином уже существует"})
		return
	}

	// Создаем нового пользователя
	user := ds.Users{
		Login:       request.Login,
		Password:    request.Password, // В реальном приложении нужно хэшировать!
		IsModerator: false,
	}

	if err := h.Repository.GetDB().Create(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Не возвращаем пароль
	user.Password = ""

	ctx.JSON(http.StatusCreated, user)
}

// GetUserProfile - GET полей пользователя
func (h *Handler) GetUserProfile(ctx *gin.Context) {
	// В этой лабораторной используем системного пользователя
	user := h.Repository.GetSystemUser()

	// Не возвращаем пароль
	user.Password = ""

	ctx.JSON(http.StatusOK, user)
}

// UpdateUserProfile - PUT пользователя
func (h *Handler) UpdateUserProfile(ctx *gin.Context) {
	var request struct {
		Login string `json:"login"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// В этой лабораторной обновляем системного пользователя
	// В реальном приложении нужно брать пользователя из сессии/токена
	user := h.Repository.GetSystemUser()

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

// LoginUser - POST аутентификация (для 4 лабораторной)
func (h *Handler) LoginUser(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "Аутентификация будет в 4 лабораторной"})
}

// LogoutUser - POST деавторизация (для 4 лабораторной)
func (h *Handler) LogoutUser(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "Деавторизация будет в 4 лабораторной"})
}
