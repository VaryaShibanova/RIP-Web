package handler

import (
	"RIP-WEB/internal/app/ds"
	"RIP-WEB/internal/app/minio"
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ДОМЕН УСЛУГИ (ANOMALY)

// GetAnomalies - GET список с фильтрацией
func (h *Handler) GetAnomalies(ctx *gin.Context) {
	var anomalies []ds.Anomaly

	// Фильтрация по названию
	name := ctx.Query("name")
	year := ctx.Query("year")

	if name != "" || year != "" {
		// Поиск с фильтрацией
		query := ""
		if name != "" {
			query = name
		}
		anomalies, _ = h.Repository.SearchAnomalies(query)

		// Дополнительная фильтрация по году если нужно
		if year != "" {
			filtered := []ds.Anomaly{}
			yearInt, _ := strconv.Atoi(year)
			for _, anomaly := range anomalies {
				if anomaly.Year == yearInt {
					filtered = append(filtered, anomaly)
				}
			}
			anomalies = filtered
		}
	} else {
		// Все аномалии
		anomalies, _ = h.Repository.GetAllAnomalies()
	}

	ctx.JSON(http.StatusOK, gin.H{
		"anomalies": anomalies,
	})
}

// GetAnomaly - GET одна запись
func (h *Handler) GetAnomaly(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	anomaly, err := h.Repository.GetAnomalyByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if anomaly == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Аномалия не найдена"})
		return
	}

	ctx.JSON(http.StatusOK, anomaly)
}

// CreateAnomaly - POST добавление (без изображения)
func (h *Handler) CreateAnomaly(ctx *gin.Context) {
	var anomaly ds.Anomaly
	if err := ctx.ShouldBindJSON(&anomaly); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Repository.CreateAnomaly(&anomaly); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, anomaly)
}

// UpdateAnomaly - PUT изменение
func (h *Handler) UpdateAnomaly(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	anomaly, err := h.Repository.GetAnomalyByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if anomaly == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Аномалия не найдена"})
		return
	}

	var updateData ds.Anomaly
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем только разрешенные поля
	anomaly.Name = updateData.Name
	anomaly.Description = updateData.Description
	anomaly.Year = updateData.Year

	if err := h.Repository.UpdateAnomaly(anomaly); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, anomaly)
}

// DeleteAnomaly - DELETE удаление
func (h *Handler) DeleteAnomaly(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	// Получаем аномалию для удаления изображения
	anomaly, err := h.Repository.GetAnomalyByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if anomaly == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Аномалия не найдена"})
		return
	}

	// Удаляем изображение из Minio если есть
	if anomaly.Image != "" {
		objectName := minio.ExtractObjectNameFromURL(anomaly.Image)
		minioClient, _ := minio.InitMinio()
		if minioClient != nil {
			minio.DeleteObject(context.Background(), minioClient, "images", objectName)
		}
	}

	if err := h.Repository.DeleteAnomaly(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Аномалия удалена"})
}

// UploadAnomalyImage - POST добавление изображения
func (h *Handler) UploadAnomalyImage(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	anomaly, err := h.Repository.GetAnomalyByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if anomaly == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Аномалия не найдена"})
		return
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Файл не найден"})
		return
	}

	// Удаляем старое изображение если есть
	if anomaly.Image != "" {
		objectName := minio.ExtractObjectNameFromURL(anomaly.Image)
		minioClient, _ := minio.InitMinio()
		if minioClient != nil {
			minio.DeleteObject(context.Background(), minioClient, "images", objectName)
		}
	}

	// Загружаем новое изображение в Minio
	minioClient, err := minio.InitMinio()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка Minio: " + err.Error()})
		return
	}

	objectName, err := minio.UploadImage(context.Background(), minioClient, "images", file, uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки: " + err.Error()})
		return
	}

	// Сохраняем URL изображения в БД
	imageURL := minio.GetImageURL(objectName)
	if err := h.Repository.UpdateAnomalyImage(uint(id), imageURL); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "Изображение загружено",
		"image_url": imageURL,
	})
}

// ДОМЕН ЗАЯВКИ (TREE)

// GetTreeCart - GET иконки корзины
func (h *Handler) GetTreeCart(ctx *gin.Context) {
	user := h.Repository.GetSystemUser()

	tree, err := h.Repository.GetDraftTree(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var treeID uint = 0
	var count int64 = 0

	if tree != nil {
		treeID = tree.ID
		count = h.Repository.GetCartCount(user.ID)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"tree_id":    treeID,
		"item_count": count,
	})
}

// GetTrees - GET список заявок с фильтрацией
func (h *Handler) GetTrees(ctx *gin.Context) {
	status := ctx.Query("status")
	dateFromStr := ctx.Query("date_from")
	dateToStr := ctx.Query("date_to")

	var dateFrom, dateTo time.Time
	var err error

	if dateFromStr != "" {
		dateFrom, err = time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты от"})
			return
		}
	}

	if dateToStr != "" {
		dateTo, err = time.Parse("2006-01-02", dateToStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты до"})
			return
		}
	}

	trees, err := h.Repository.GetTreesWithFilters(status, dateFrom, dateTo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Преобразуем для клиента (только логины)
	type TreeResponse struct {
		ID          uint      `json:"id"`
		Status      string    `json:"status"`
		Description string    `json:"description"`
		TotalRings  int       `json:"total_rings"`
		FinalYear   int       `json:"final_year"`
		DateCreate  time.Time `json:"date_create"`
		DateUpdate  time.Time `json:"date_update"`
		DateFinish  time.Time `json:"date_finish,omitempty"`
		Creator     string    `json:"creator"`
		Moderator   string    `json:"moderator,omitempty"`
	}

	response := make([]TreeResponse, len(trees))
	for i, tree := range trees {
		response[i] = TreeResponse{
			ID:          tree.ID,
			Status:      tree.Status,
			Description: tree.Description,
			TotalRings:  tree.TotalRings,
			FinalYear:   tree.FinalYear,
			DateCreate:  tree.DateCreate,
			DateUpdate:  tree.DateUpdate,
			Creator:     tree.Creator.Login,
		}

		if tree.ModeratorID.Valid {
			response[i].Moderator = tree.Moderator.Login
		}
		if tree.DateFinish.Valid {
			response[i].DateFinish = tree.DateFinish.Time
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"trees": response,
	})
}

// GetTree - GET одна запись заявки
func (h *Handler) GetTree(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	tree, treeItems, err := h.Repository.GetTreeWithItems(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Проверяем что заявка не удалена
	if tree.Status == "удалён" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка удалена"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"tree":      tree,
		"treeItems": treeItems,
	})
}

// UpdateTree - PUT изменения полей заявки
func (h *Handler) UpdateTree(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	tree, err := h.Repository.GetTreeByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var updateData struct {
		Description string `json:"description"`
		TotalRings  int    `json:"total_rings"`
	}

	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем только разрешенные поля
	tree.Description = updateData.Description
	tree.TotalRings = updateData.TotalRings

	if err := h.Repository.UpdateTree(tree); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, tree)
}

// FormTree - PUT сформировать заявку
func (h *Handler) FormTree(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	tree, err := h.Repository.GetTreeByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Проверяем что заявка в статусе черновика
	if tree.Status != "черновик" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Можно формировать только черновые заявки"})
		return
	}

	if err := h.Repository.FormTree(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем обновленную заявку
	updatedTree, err := h.Repository.GetTreeByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedTree)
}

// CompleteTree - PUT завершить/отклонить заявку
func (h *Handler) CompleteTree(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	var request struct {
		Action string `json:"action" binding:"required"` // "complete" или "reject"
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Action != "complete" && request.Action != "reject" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Действие должно быть 'complete' или 'reject'"})
		return
	}

	moderator := h.Repository.GetModerator()

	if err := h.Repository.CompleteTree(uint(id), moderator.ID, request.Action); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем обновленную заявку
	updatedTree, err := h.Repository.GetTreeByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedTree)
}

// DeleteTree - DELETE удаление заявки
func (h *Handler) DeleteTree(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	if err := h.Repository.DeleteTree(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Заявка удалена"})
}

// ДОМЕН М-М (TREE ITEM)

// AddToTree - POST добавление в заявку-черновик
func (h *Handler) AddToTree(ctx *gin.Context) {
	user := h.Repository.GetSystemUser()

	var request struct {
		AnomalyID uint `json:"anomaly_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем существование аномалии
	anomaly, err := h.Repository.GetAnomalyByID(int(request.AnomalyID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if anomaly == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Аномалия не найдена"})
		return
	}

	// Создаем или получаем черновую заявку
	tree, err := h.Repository.GetOrCreateDraftTree(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Добавляем аномалию в заявку
	err = h.Repository.AddAnomalyToTree(tree.ID, request.AnomalyID, "", 0)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Аномалия добавлена в заявку",
		"tree_id": tree.ID,
	})
}

// UpdateTreeItem - PUT изменение значений в м-м
func (h *Handler) UpdateTreeItem(ctx *gin.Context) {
	treeID, err := strconv.Atoi(ctx.Param("tree_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID заявки"})
		return
	}

	anomalyID, err := strconv.Atoi(ctx.Param("anomaly_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID аномалии"})
		return
	}

	var request struct {
		AnomalousRings string `json:"anomalous_rings"`
		CalculatedYear int    `json:"calculated_year"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Repository.UpdateTreeItem(uint(treeID), uint(anomalyID), request.AnomalousRings, request.CalculatedYear); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Элемент заявки не найден"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Элемент заявки обновлен"})
}

// RemoveFromTree - DELETE удаление из заявки
func (h *Handler) RemoveFromTree(ctx *gin.Context) {
	treeID, err := strconv.Atoi(ctx.Param("tree_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID заявки"})
		return
	}

	anomalyID, err := strconv.Atoi(ctx.Param("anomaly_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID аномалии"})
		return
	}

	if err := h.Repository.RemoveFromTree(uint(treeID), uint(anomalyID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Элемент удален из заявки"})
}

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
	err := h.db.Where("login = ?", request.Login).First(&existingUser).Error
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

	if err := h.db.Create(&user).Error; err != nil {
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
		err := h.db.Where("login = ? AND id != ?", request.Login, user.ID).First(&existingUser).Error
		if err == nil {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Логин уже занят"})
			return
		}
		user.Login = request.Login
	}

	if err := h.db.Save(user).Error; err != nil {
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
