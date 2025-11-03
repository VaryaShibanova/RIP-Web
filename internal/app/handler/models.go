package handler

// AnomaliesListResponse представляет ответ со списком аномалий
type AnomaliesListResponse struct {
	Anomalies []AnomalyShortResponse `json:"anomalies"`
}

// AnomalyShortResponse представляет краткую информацию об аномалии
type AnomalyShortResponse struct {
	ID       uint   `json:"id" example:"1"`
	Name     string `json:"name" example:"Аномалия роста"`
	ImageURL string `json:"image_url" example:"http://localhost:9000/images/anomaly_1.jpg"`
	Year     int    `json:"year" example:"2023"`
}

// AnomalyDetailResponse представляет полную информацию об аномалии
type AnomalyDetailResponse struct {
	ID          uint   `json:"id" example:"1"`
	Name        string `json:"name" example:"Аномалия роста"`
	Description string `json:"description" example:"Описание аномалии"`
	ImageURL    string `json:"image_url" example:"http://localhost:9000/images/anomaly_1.jpg"`
	Year        int    `json:"year" example:"2023"`
}

// CreateAnomalyRequest представляет запрос на создание аномалии
type CreateAnomalyRequest struct {
	Name        string `json:"name" binding:"required" example:"Аномалия роста"`
	Description string `json:"description" binding:"required" example:"Описание аномалии"`
	Year        int    `json:"year" binding:"required" example:"2023"`
}

// UpdateAnomalyRequest представляет запрос на обновление аномалии
type UpdateAnomalyRequest struct {
	Name        string `json:"name" example:"Обновленное название"`
	Description string `json:"description" example:"Обновленное описание"`
	Year        int    `json:"year" example:"2024"`
}

// UpdateAnomalyResponse представляет ответ на обновление аномалии
type UpdateAnomalyResponse struct {
	Message string                `json:"message" example:"Информация об аномалии обновлена"`
	Anomaly AnomalyDetailResponse `json:"anomaly"`
}

// UploadImageResponse представляет ответ на загрузку изображения
type UploadImageResponse struct {
	Message  string `json:"message" example:"Изображение загружено"`
	ImageURL string `json:"image_url" example:"http://localhost:9000/images/anomaly_1.jpg"`
	Filename string `json:"filename" example:"anomaly_image"`
}

// TreeCartPublicResponse представляет публичный ответ с данными корзины (без авторизации)
type TreeCartPublicResponse struct {
	UserID    int64 `json:"user_id" example:"-1"`
	ItemCount int64 `json:"item_count" example:"0"`
}

// TreeCartResponse представляет ответ с данными корзины - добаыить user_id
type TreeCartResponse struct {
	TreeID    uint  `json:"tree_id" example:"1"`
	ItemCount int64 `json:"item_count" example:"5"`
}

// TreesListResponse представляет ответ со списком заявок
type TreesListResponse struct {
	Trees []TreeShortResponse `json:"trees"`
}

// TreeShortResponse представляет краткую информацию о заявке
type TreeShortResponse struct {
	ID                uint   `json:"id" example:"1"`
	Creator           string `json:"creator" example:"research_user"`
	Moderator         string `json:"moderator,omitempty" example:"moderator_user"`
	AmountOfAnomalies int    `json:"amount_of_anomalies" example:"3"`
	Status            string `json:"status,omitempty" example:"черновик"` // Добавляем для модератора
}

// TreeResponse представляет информацию о заявке
type TreeResponse struct {
	ID          uint   `json:"id" example:"1"`
	Description string `json:"description" example:"Описание заявки"`
	TotalRings  int    `json:"total_rings" example:"100"`
	FinalYear   int    `json:"final_year" example:"2023"`
	Status      string `json:"status,omitempty" example:"черновик"` // Добавляем для модератора
	CreatorID   uint   `json:"creator_id" example:"1"`
}

// TreeDetailResponse представляет полную информацию о заявке
type TreeDetailResponse struct {
	Tree      TreeResponse       `json:"tree"`
	TreeItems []TreeItemResponse `json:"treeItems"`
}

// TreeItemResponse представляет информацию об элементе заявки
type TreeItemResponse struct {
	AnomalyID      uint   `json:"anomaly_id" example:"1"`
	AnomalousRings string `json:"anomalous_rings" example:"45,67,89"`
	CalculatedYear int    `json:"calculated_year" example:"2023"`
	AnomalyName    string `json:"anomaly_name" example:"Аномалия роста"`
	AnomalyImage   string `json:"anomaly_image" example:"http://localhost:9000/images/anomaly_1.jpg"`
}

// UpdateTreeRequest представляет запрос на обновление заявки
type UpdateTreeRequest struct {
	Description string `json:"description" example:"Обновленное описание"`
	TotalRings  int    `json:"total_rings" example:"120"`
	FinalYear   int    `json:"final_year" example:"2024"`
}

// CompleteTreeRequest представляет запрос на завершение заявки
type CompleteTreeRequest struct {
	Action string `json:"action" binding:"required" example:"complete"`
}

// AddToTreeRequest представляет запрос на добавление в заявку
type AddToTreeRequest struct {
	AnomalyID uint `json:"anomaly_id" binding:"required" example:"1"`
}

// AddToTreeResponse представляет ответ на добавление в заявку
type AddToTreeResponse struct {
	Message string `json:"message" example:"Аномалия добавлена в заявку"`
	TreeID  uint   `json:"tree_id" example:"1"`
}

// UpdateTreeItemRequest представляет запрос на обновление элемента заявки
type UpdateTreeItemRequest struct {
	AnomalousRings string `json:"anomalous_rings" example:"45,67,89"`
}

// UpdateTreeItemResponse представляет ответ на обновление элемента заявки
type UpdateTreeItemResponse struct {
	Message        string `json:"message" example:"Элемент заявки обновлен"`
	AnomalousRings string `json:"anomalous_rings" example:"45,67,89"`
	CalculatedYear int    `json:"calculated_year" example:"0"`
}

// CompleteTreeResponse представляет ответ на завершение заявки
type CompleteTreeResponse struct {
	ID             uint                    `json:"id" example:"1"`
	Status         string                  `json:"status" example:"завершён"`
	FinalYear      int                     `json:"final_year" example:"2023"`
	Anomalies      []AnomalyCalculatedYear `json:"anomalies"`
	TotalAnomalies int                     `json:"total_anomalies" example:"3"`
}

// AnomalyCalculatedYear представляет аномалию с рассчитанным годом
type AnomalyCalculatedYear struct {
	AnomalyID      uint   `json:"anomaly_id" example:"1"`
	AnomalyName    string `json:"anomaly_name" example:"Аномалия роста"`
	AnomalousRings string `json:"anomalous_rings" example:"45,67,89"`
	CalculatedYear int    `json:"calculated_year" example:"2023"`
}
