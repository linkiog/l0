package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/linkiog/lo/internal/cache"
	"github.com/linkiog/lo/internal/config"
	"github.com/linkiog/lo/internal/repository"
)

func StartConsumerGroup(ctx context.Context, cfg *config.Config, repo *repository.Repository, c *cache.Cache) error {
	consumer := NewConsumer(repo, c)

	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	client, err := sarama.NewConsumerGroup([]string{cfg.KafkaBrokers}, "my-group", saramaConfig)
	if err != nil {
		return err
	}

	go func() {
		defer func() {
			if err := client.Close(); err != nil {
				fmt.Println("Error closing client:", err)
			}
		}()

		for {
			if err := client.Consume(ctx, []string{cfg.KafkaTopic}, consumer); err != nil {
				fmt.Println("Error from consumer:", err)
				time.Sleep(time.Second)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	<-consumer.ready
	fmt.Println("Kafka consumer up and running!")
	return nil
}
