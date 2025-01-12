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

func NewConsumer(repo *repository.Repository, c *cache.Cache) *Consumer {
	return &Consumer{
		ready: make(chan bool),
		repo:  repo,
		cache: c,
	}
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	for message := range claim.Messages() {
		fmt.Printf("Message claimed: key = %s, value = %s, topic = %s\n",
			string(message.Key), string(message.Value), message.Topic)

		if len(message.Value) == 0 {
			fmt.Println("Received an empty message, skipping...")
			continue
		}

		var order models.Order
		err := json.Unmarshal(message.Value, &order)
		if err != nil {
			fmt.Printf("Error unmarshalling JSON: %v, message: %s\n", err, string(message.Value))
			continue
		}

		fmt.Printf("Parsed order: %+v\n", order)

		err = c.repo.SaveOrder(&order)
		if err != nil {
			fmt.Printf("Error saving order to DB: %v\n", err)
			continue
		}

		fmt.Printf("Order saved to DB successfully: %s\n", order.OrderUID)

		c.cache.Set(&order)
		fmt.Printf("Order cached successfully: %s\n", order.OrderUID)

		session.MarkMessage(message, "")
		fmt.Printf("Message offset marked as read: %d\n", message.Offset)
	}

	return nil
}
