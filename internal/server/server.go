package server

import (
	"encoding/json"

	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkiog/lo/internal/cache"
	"github.com/linkiog/lo/internal/repository"
)

func NewServer(repo *repository.Repository, c *cache.Cache) http.Handler {
	r := chi.NewRouter()

	r.Get("/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		orderUID := chi.URLParam(r, "id")

		// Пытаемся взять из кэша
		if order, ok := c.Get(orderUID); ok {
			writeJSON(w, order)
			return
		}

		// Если в кэше нет, можно сходить в БД
		order, err := repo.GetOrder(orderUID)
		if err != nil {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		// Сохраним в кэш, чтобы в след. раз было быстрее
		c.Set(order)
		writeJSON(w, order)
	})

	return r
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
