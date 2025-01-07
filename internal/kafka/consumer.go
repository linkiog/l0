package kafka

import (
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/linkiog/lo/internal/cache"
	"github.com/linkiog/lo/internal/models"
	"github.com/linkiog/lo/internal/repository"
)

type Consumer struct {
	ready chan bool
	repo  *repository.Repository
	cache *cache.Cache
}

// NewConsumer создаёт Consumer, которому нужны репозиторий (для записи в БД) и кэш (для хранения в памяти).
func NewConsumer(repo *repository.Repository, c *cache.Cache) *Consumer {
	return &Consumer{
		ready: make(chan bool),
		repo:  repo,
		cache: c,
	}
}

// Setup вызывается перед тем, как потребитель начнёт читать партиции
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

// Cleanup вызывается в конце
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim - основная логика чтения из топика
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	for message := range claim.Messages() {
		fmt.Printf("Message claimed: key = %s, value = %s, topic = %s\n",
			string(message.Key), string(message.Value), message.Topic)

		var order models.Order
		err := json.Unmarshal(message.Value, &order)
		if err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			// Можно отправить в dead-letter-queue, либо пропустить
			continue
		}

		// Сохраняем в БД
		if err := c.repo.SaveOrder(&order); err != nil {
			fmt.Println("Error saving to DB:", err)
			continue
		}

		// В кэш
		c.cache.Set(&order)

		// Помечаем offset как прочитанный
		session.MarkMessage(message, "")
	}

	return nil
}
