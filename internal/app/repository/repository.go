package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Anomaly struct {
	ID          int
	Year        int
	Pattern     string
	Description string
	RingNumbers string
	Image       string
}

type Request struct {
	ID              int
	AnomalyName     string
	FindDescription string
	TotalRings      int
	AnomalousRings  string
	CalculatedYear  int
	Image           string
}

func (r *Repository) GetAnomalies() ([]Anomaly, error) {
	anomalies := []Anomaly{
		{
			ID:          1,
			Year:        1600,
			Pattern:     "Извержение вулкана Уайнапутина",
			Description: "На срезе дерева, жившего в 1600 году, хорошо заметно одно очень узкое и темное годовое кольцо. Оно резко контрастирует с более широкими светлыми кольцами до и после него. Это кольцо 1601 года.",
			Image:       "http://127.0.0.1:9000/images/img/card1.jpg",
		},
		{
			ID:          2,
			Year:        1815,
			Pattern:     "Извержение Тамбора",
			Description: "На срезе дерева хорошо видно три аномальных кольца старше. Такие кольца 1815 года характеризуются узкими, фрагментированными темными кольцами 1816 года и относительно широкими кольцами 1817 года.",
			Image:       "http://127.0.0.1:9000/images/img/card2.jpg",
		},
		{
			ID:          3,
			Year:        1883,
			Pattern:     "Извержение Каракатау",
			Description: "На срезе дерева, примерно на 15-16 кольцах от края, видны аномальные кольца второго следования 1884 года, например фрагментированное кольцо 1883 года и относительно широкими кольцами 1882 года.",
			Image:       "http://127.0.0.1:9000/images/img/card3.jpg",
		},
		{
			ID:          4,
			Year:        1931,
			Pattern:     "Великое наводнение в Китае",
			Description: "Примерно на 50-90 кольцах от края видна аномальная необычная ширина, начинающаяся около 1931 года. Оно резко контрастирует с более узкими и темными кольцами предыдущих лет.",
			Image:       "http://127.0.0.1:9000/images/img/card4.jpg",
		},
		{
			ID:          5,
			Year:        1962,
			Pattern:     "Наводнение Бурхарди",
			Description: "На срезе дерева, росшего в тот период, видно редкое очень широкое, с элементами черных волокон, кольцо соответствующее 1962 году или первому году после события. Оно заметно выделяется на фоне колец обычной ширины.",
			Image:       "http://127.0.0.1:9000/images/img/card5.jpg",
		},
		{
			ID:          6,
			Year:        1816,
			Pattern:     "Год без лета",
			Description: "Чёткая тёмная полоса 1816 года — «год без лета» — выглядит как шрам, врезавшийся в память дерева. Это сверхузкое, почти чёрное кольцо. Рядом видны более светлые и широкие кольца.",
			Image:       "http://127.0.0.1:9000/images/img/card6.jpg",
		},
		{
			ID:          7,
			Year:        1657,
			Pattern:     "Великий пожар годов Мэйрэки",
			Description: "На срезе вида аномалия на 15-м кольце от коры. Само кольцо 1657 года неровное и фрагментированное: с одной стороны узкое и плотное, с другой — более широкое. За ним следует аномально широкое светлое кольцо 1658 года.",
			Image:       "http://127.0.0.1:9000/images/img/card7.jpg",
		},
	}

	if len(anomalies) == 0 {
		return nil, fmt.Errorf("нет данных об аномалиях")
	}

	return anomalies, nil
}

func (r *Repository) GetAnomaly(id int) (Anomaly, error) {
	anomalies, err := r.GetAnomalies()
	if err != nil {
		return Anomaly{}, err
	}

	for _, anomaly := range anomalies {
		if anomaly.ID == id {
			return anomaly, nil
		}
	}
	return Anomaly{}, fmt.Errorf("аномалия не найдена")
}

func (r *Repository) GetAnomaliesByPattern(pattern string) ([]Anomaly, error) {
	anomalies, err := r.GetAnomalies()
	if err != nil {
		return []Anomaly{}, err
	}

	var result []Anomaly
	for _, anomaly := range anomalies {
		if strings.Contains(strings.ToLower(anomaly.Pattern), strings.ToLower(pattern)) ||
			strings.Contains(strings.ToLower(anomaly.Description), strings.ToLower(pattern)) {
			result = append(result, anomaly)
		}
	}

	return result, nil
}

func (r *Repository) GetRequests() ([]Request, error) {
	requests := []Request{
		{
			ID:              1,
			AnomalyName:     "Извержение вулкана Уайнапутина",
			FindDescription: "Обнаружено узкое темное кольцо 1601 года",
			TotalRings:      320,
			AnomalousRings:  "1601",
			CalculatedYear:  1600,
			Image:           "http://127.0.0.1:9000/images/img/card1.jpg",
		},
		{
			ID:              2,
			AnomalyName:     "Извержение Каракатау",
			FindDescription: "Обнаружены фрагментированные кольца 1883-1884 годов",
			TotalRings:      150,
			AnomalousRings:  "1883, 1884",
			CalculatedYear:  1883,
			Image:           "http://127.0.0.1:9000/images/img/card3.jpg",
		},
	}

	if len(requests) == 0 {
		return nil, fmt.Errorf("нет данных о заявках")
	}

	return requests, nil
}
