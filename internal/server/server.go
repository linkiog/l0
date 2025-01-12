package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/linkiog/lo/internal/cache"
	"github.com/linkiog/lo/internal/repository"
)

func NewServer(repo *repository.Repository, c *cache.Cache) http.Handler {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		html := `
        <!DOCTYPE html>
        <html>
        <head>
            <meta charset="utf-8"/>
            <title>Order Viewer</title>
        </head>
        <body>
            <h1>Order Viewer (JSON)</h1>
            <label for="orderId">Введите ID заказа:</label>
            <input type="text" id="orderId" placeholder="Пример: 12345">
            <button onclick="fetchOrder()">Получить заказ</button>
            <pre id="result"></pre>

            <script>
                function fetchOrder() {
                    const id = document.getElementById('orderId').value;
                    // Запрос к эндпоинту /orders/{id}, который возвращает JSON
                    fetch('/orders/' + id)
                        .then(response => {
                            if (!response.ok) {
                                throw new Error('Order not found or invalid ID');
                            }
                            return response.json();
                        })
                        .then(data => {
                            document.getElementById('result').textContent = JSON.stringify(data, null, 2);
                        })
                        .catch(error => {
                            document.getElementById('result').textContent = 'Ошибка: ' + error.message;
                        });
                }
            </script>
        </body>
        </html>
        `
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(html))
	})

	r.Get("/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		orderUID := chi.URLParam(r, "id")
		if orderUID == "" {
			http.Error(w, "missing or invalid order ID", http.StatusBadRequest)
			return
		}

		if order, ok := c.Get(orderUID); ok {
			fmt.Printf("Order %s returned from cache\n", orderUID)
			writeJSON(w, order)
			return
		}

		fmt.Printf("Order %s not found in cache, querying DB...\n", orderUID)
		order, err := repo.GetOrder(orderUID)
		if err != nil {
			http.Error(w, "order not found", http.StatusNotFound)
			log.Printf("error fetching order %s from DB: %v", orderUID, err)
			return
		}
		if order == nil {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}

		c.Set(order)
		fmt.Printf("Order %s loaded from DB and cached\n", orderUID)

		writeJSON(w, order)
	})

	return r
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}
