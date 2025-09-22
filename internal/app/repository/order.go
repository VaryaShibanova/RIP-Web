package repository

import (
	"fmt"

	"RIP-WEB/internal/app/ds"
)

func (r *Repository) GetOrders() ([]ds.Order, error) {
	var orders []ds.Order
	err := r.db.Find(&orders).Error
	// обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return orders, nil
}

func (r *Repository) GetOrder(id int) (ds.Order, error) {
	order := ds.Order{}
	err := r.db.Where("id = ?", id).First(&order).Error
	if err != nil {
		return ds.Order{}, err
	}
	return order, nil
}

func (r *Repository) GetOrdersByTitle(title string) ([]ds.Order, error) {
	var orders []ds.Order
	err := r.db.Where("name ILIKE ?", "%"+title+"%").Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}
