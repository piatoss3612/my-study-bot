package rabbitmq

import (
	"errors"

	"github.com/piatoss3612/my-study-bot/internal/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
)

var ErrMissingXEventTopicHeader = errors.New("missing x-event-topic header")

type subscriber struct {
	conn     *amqp.Connection
	exchange string
	queue    string
}

func NewSubscriber(conn *amqp.Connection, exchange, kind, queue string) (pubsub.Subscriber, error) {
	sub := &subscriber{
		conn:     conn,
		exchange: exchange,
		queue:    queue,
	}

	return sub.setup(exchange, kind, queue)
}

func (s *subscriber) setup(exchange, kind, queue string) (pubsub.Subscriber, error) {
	ch, err := s.conn.Channel()
	if err != nil {
		return nil, err
	}
	defer func() { _ = ch.Close() }()

	err = ch.ExchangeDeclare(exchange, kind, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *subscriber) Subscribe(topics ...string) (<-chan pubsub.Message, <-chan error, func(), error) {
	ch, err := s.conn.Channel()
	if err != nil {
		return nil, nil, nil, err
	}

	for _, topic := range topics {
		if err := ch.QueueBind(s.queue, topic, s.exchange, false, nil); err != nil {
			return nil, nil, nil, err
		}
	}

	delivery, err := ch.Consume(s.queue, "", false, false, false, false, nil)
	if err != nil {
		return nil, nil, nil, err
	}

	msgs := make(chan pubsub.Message)
	errs := make(chan error)

	go s.handleMessage(delivery, msgs, errs)

	return msgs, errs, func() {
		_ = ch.Close()
		close(msgs)
		close(errs)
	}, nil
}

func (s *subscriber) handleMessage(delivery <-chan amqp.Delivery, msgs chan<- pubsub.Message, errs chan<- error) {
	for d := range delivery {
		topic, ok := d.Headers[XEventTopicHeader]
		if !ok {
			errs <- ErrMissingXEventTopicHeader
			_ = d.Nack(false, false)
			continue
		}

		msg := pubsub.Message{
			Topic: topic.(string),
			Body:  d.Body,
		}

		msgs <- msg
		_ = d.Ack(false)
	}
}
