package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("Could not create channel: %v", err)
	}

	durable := simpleQueueType == SimpleQueueDurable
	autoDelete := simpleQueueType == SimpleQueueTransient
	exclusive := simpleQueueType == SimpleQueueTransient
	queue, err := channel.QueueDeclare(
		queueName,
		durable,
		autoDelete,
		exclusive,
		false,
		amqp.Table{"x-dead-letter-exchange": "peril_dlx"},
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("Could not declare queue: %v", err)
	}

	err = channel.QueueBind(queue.Name, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("Could not bind queue: %v", err)
	}
	return channel, queue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {
	return subscribe(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
		handler,
		func(data []byte) (T, error) {
			var unmarshalled T
			err := json.Unmarshal(data, &unmarshalled)
			return unmarshalled, err
		},
	)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {
	return subscribe(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
		handler,
		func(data []byte) (T, error) {
			buffer := bytes.NewBuffer(data)
			dec := gob.NewDecoder(buffer)
			var decoded T
			err := dec.Decode(&decoded)
			return decoded, err
		},
	)
}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return fmt.Errorf("could not declare and bind queue: %v", err)
	}

	err = channel.Qos(10, 0, false)
	if err != nil {
		return fmt.Errorf("could not set QoS: %v", err)
	}

	msgs, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not consume messages: %v", err)
	}
	go func() {
		defer channel.Close()
		for msg := range msgs {
			data, err := unmarshaller(msg.Body)
			if err != nil {
				fmt.Printf("could not unmarshall the message: %v\n", err)
				continue
			}

			switch handler(data) {
			case Ack:
				fmt.Println("Acknowledging message")
				err = msg.Ack(false)
			case NackRequeue:
				fmt.Println("Requeuing message")
				err = msg.Nack(false, true)
			case NackDiscard:
				fmt.Println("Discarding message")
				err = msg.Nack(false, false)
			}

			if err != nil {
				fmt.Printf("could not acknowledge message: %v\n", err)
			}
		}
	}()
	return nil
}
