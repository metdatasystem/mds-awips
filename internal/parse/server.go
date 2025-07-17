package parse

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metdatasystem/mds-awips/internal/parse/handler"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	Exchange   = "awips.exchange"
	Queue      = "awips.queue"
	ParseRoute = "awips.parse"
)

type Config struct {
	MinLog int
}

type Server struct {
	DB        *pgxpool.Pool
	Rabbit    *amqp.Connection
	Consumer  *amqp.Channel
	Publisher *amqp.Channel
	Config    Config
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

func New(config Config) (*Server, error) {
	// Create a new database connection pool
	db, err := db.New()
	if err != nil {
		return nil, err
	}

	rabbitConn, err := rabbit.NewConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %s", err.Error())
	}

	consumer, err := rabbit.NewConsumerChannel(rabbitConn, Queue, Exchange, ParseRoute)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer channel: %v", err)
	}

	publisher, err := rabbit.NewPublisherChannel(rabbitConn, Exchange, "topic")
	if err != nil {
		return nil, fmt.Errorf("failed to crete publisher channel: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	server := Server{
		DB:        db,
		Rabbit:    rabbitConn,
		Config:    config,
		Consumer:  consumer,
		Publisher: publisher,
		ctx:       ctx,
		cancel:    cancel,
	}

	return &server, nil
}

type Message struct {
	Text       string
	ReceivedAt time.Time
}

func (server *Server) Start() {

	// Perform health check, not sure if this does anything useful
	err := server.HeathCheck()
	if err != nil {
		slog.Error(err.Error())
		slog.Info("Stopping...")
		return
	}

	msgs, err := server.Consumer.Consume(
		Queue, // queue
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		slog.Error("failed to register consumer: " + err.Error())
		return
	}

	slog.Info("\033[32m *** Consumer listening *** \033[m")

	// Listen for messages and process them
	go func() {
		for message := range msgs {
			m := &Message{}
			err := json.Unmarshal(message.Body, m)
			if err != nil {
				slog.Error(err.Error())
				return
			}

			server.wg.Add(1)
			go func(text string, receivedAt time.Time) {
				defer server.wg.Done() // Decrement when done
				h := handler.New(server.DB, server.Config.MinLog, text, receivedAt)
				h.Handle()
				err := h.SaveLog()
				if err != nil {
					slog.Error("failed to save log", "error", err)
				}
			}(m.Text, m.ReceivedAt.UTC())
		}
	}()
}

func (server *Server) Shutdown() {
	server.cancel()
	server.wg.Wait()
	err := server.Consumer.Close()
	if err != nil {
		slog.Error("failed to close consumer", "error", err)
	}
	err = server.Publisher.Close()
	if err != nil {
		slog.Error("failed to close publisher", "error", err)
	}
	err = server.Rabbit.Close()
	if err != nil {
		slog.Error("failed to close RabbitMQ connection", "error", err)
	}
	server.DB.Close()
}

func (server *Server) HeathCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.DB.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}
