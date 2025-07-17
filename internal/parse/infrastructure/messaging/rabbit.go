package rabbit

import (
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

func NewConnection() (*amqp.Connection, error) {

	conn, err := amqp.Dial(os.Getenv("RABBIT"))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func NewConsumerChannel(conn *amqp.Connection, queueName string, exchange string, routingKey string) (*amqp.Channel, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, err
	}

	err = channel.QueueBind(
		queue.Name, // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,
		nil)
	if err != nil {
		return nil, err
	}

	return channel, nil
}

func NewPublisherChannel(conn *amqp.Connection, exchangeName, exchangeType string) (*amqp.Channel, error) {

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(
		exchangeName, // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return nil, err
	}

	return ch, nil

}
