package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/streadway/amqp"

	"github.com/paradoxe35/torra/packages/config"
	"github.com/paradoxe35/torra/packages/logger"
)

// RabbitMQAdapter implements the EventPublisher interface using RabbitMQ
type RabbitMQAdapter struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	exchange     string
	routingKey   string
	logger       *logger.Logger
	publishQueue chan publishRequest
	done         chan bool
}

type publishRequest struct {
	ctx   context.Context
	event interface{}
	done  chan error
}

// NewRabbitMQEventPublisher creates a new RabbitMQ event publisher
func NewRabbitMQEventPublisher(cfg *config.RabbitMQConfig, logger *logger.Logger) (EventPublisher, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		cfg.Exchange, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	adapter := &RabbitMQAdapter{
		conn:         conn,
		channel:      channel,
		exchange:     cfg.Exchange,
		routingKey:   cfg.Queue, // Use Queue as routing key
		logger:       logger,
		publishQueue: make(chan publishRequest, 1000),
		done:         make(chan bool),
	}

	// Start async publisher goroutine
	go adapter.asyncPublisher()

	return adapter, nil
}

// Publish publishes an event synchronously
func (r *RabbitMQAdapter) Publish(ctx context.Context, event interface{}) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Set context timeout if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	return r.channel.Publish(
		r.exchange,   // exchange
		r.routingKey, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent,
		},
	)
}

// PublishAsync publishes an event asynchronously
func (r *RabbitMQAdapter) PublishAsync(ctx context.Context, event interface{}) error {
	done := make(chan error, 1)

	select {
	case r.publishQueue <- publishRequest{ctx: ctx, event: event, done: done}:
		return <-done
	case <-ctx.Done():
		return ctx.Err()
	}
}

// PublishBatch publishes multiple events in a batch
func (r *RabbitMQAdapter) PublishBatch(ctx context.Context, events []interface{}) error {
	if len(events) == 0 {
		return nil
	}

	// Set context timeout if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
	}

	for _, event := range events {
		if err := r.Publish(ctx, event); err != nil {
			r.logger.Error(fmt.Sprintf("Failed to publish event in batch: %v", err))
			return err
		}
	}

	return nil
}

// asyncPublisher handles async publishing in a separate goroutine
func (r *RabbitMQAdapter) asyncPublisher() {
	for {
		select {
		case req := <-r.publishQueue:
			err := r.Publish(req.ctx, req.event)
			req.done <- err
		case <-r.done:
			return
		}
	}
}

// Close closes the RabbitMQ connection
func (r *RabbitMQAdapter) Close() error {
	close(r.done)

	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			r.logger.Error(fmt.Sprintf("Error closing channel: %v", err))
		}
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			r.logger.Error(fmt.Sprintf("Error closing connection: %v", err))
			return err
		}
	}

	return nil
}
