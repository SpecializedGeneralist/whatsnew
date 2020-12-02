// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rabbitmq

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

// Client provides a simple interface for connecting to RabbitMQ, declaring a
// channel, and declaring a durable Topic Exchange, which is used by various
// pub/sub operations all along the program.
type Client struct {
	amqpURI      string
	exchangeName string
	connection   *amqp.Connection
	channel      *amqp.Channel
}

// NewClient creates a new Client initialized with the given AMQP URI and
// durable Topic Exchange name.
func NewClient(amqpURI, exchangeName string) *Client {
	return &Client{
		amqpURI:      amqpURI,
		exchangeName: exchangeName,
	}
}

// Connect opens a new RabbitMQ connection to the Client's AMQP URI, then it
// opens a unique server channel and uses it to declare a durable Topic
// Exchange using the Client's Exchange name.
//
// When you are done using the client, remember to call Client.Disconnect
// to disconnect and free the resources.
func (c *Client) Connect() (err error) {
	if c.connection != nil || c.channel != nil {
		return fmt.Errorf("AMQP connection and/or channel already exist")
	}

	c.connection, err = amqp.Dial(c.amqpURI)
	if err != nil {
		return fmt.Errorf("dial RabbitMQ AMQP URI %s: %v", c.amqpURI, err)
	}
	defer func() {
		if err != nil {
			_ = c.closeConnection() // ignoring any closing error
		}
	}()

	c.channel, err = c.connection.Channel()
	if err != nil {
		return fmt.Errorf("open RabbitMQ channel: %v", err)
	}
	defer func() {
		if err != nil {
			_ = c.closeChannel() // ignoring any closing error
		}
	}()

	err = c.channel.ExchangeDeclare(
		c.exchangeName,
		amqp.ExchangeTopic,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("declare RabbitMQ exchange %s: %v", c.exchangeName, err)
	}
	return nil
}

// Disconnect closes the RabbitMQ channel and connection.
func (c *Client) Disconnect() error {
	channelClosingError := c.closeChannel()
	connectionClosingError := c.closeConnection()
	if channelClosingError != nil {
		return fmt.Errorf("close RabbitMQ channel: %v", channelClosingError)
	}
	if connectionClosingError != nil {
		return fmt.Errorf("close RabbitMQ connection: %v", connectionClosingError)
	}
	return nil
}

// PublishID publishes a new persistent message using the given routing key;
// the message consists in a JSON object containing given ID.
func (c *Client) PublishID(routingKey string, id uint) error {
	err := c.channel.Publish(
		c.exchangeName,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:     "application/json",
			ContentEncoding: "UTF-8",
			DeliveryMode:    amqp.Persistent,
			MessageId:       fmt.Sprintf("%s_%d", routingKey, id),
			Timestamp:       time.Now(),
			Body:            EncodeIDMessage(id),
		})
	if err != nil {
		return fmt.Errorf("pusblish RabbitMQ message with ID %d: %v", id, err)
	}
	return nil
}

// PublishStringID publishes a new persistent message using the given routing key;
// the message consists in a JSON object containing given string ID.
func (c *Client) PublishStringID(routingKey string, id string) error {
	err := c.channel.Publish(
		c.exchangeName,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:     "application/json",
			ContentEncoding: "UTF-8",
			DeliveryMode:    amqp.Persistent,
			MessageId:       fmt.Sprintf("%s_%s", routingKey, id),
			Timestamp:       time.Now(),
			Body:            EncodeStringIDMessage(id),
		})
	if err != nil {
		return fmt.Errorf("pusblish RabbitMQ message with ID %s: %v", id, err)
	}
	return nil
}

// Consume immediately starts delivering queued messages.
func (c *Client) Consume(queueName string, routingKey string) (<-chan amqp.Delivery, string, error) {
	queue, err := c.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, "", err
	}

	err = c.channel.QueueBind(
		queue.Name,     // queue name
		routingKey,     // routing key
		c.exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return nil, "", err
	}

	consumerTag, err := uniqueConsumerTag(queue.Name, routingKey)
	if err != nil {
		return nil, "", err
	}

	msgs, err := c.channel.Consume(
		queue.Name,  // queue
		consumerTag, // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return nil, "", err
	}
	return msgs, consumerTag, nil
}

// Cancel stops deliveries to the consumer.
func (c *Client) CancelConsumer(consumerTag string) error {
	return c.channel.Cancel(consumerTag, false)
}

func (c *Client) closeConnection() error {
	err := c.connection.Close()
	c.connection = nil
	return err
}

func (c *Client) closeChannel() error {
	err := c.channel.Close()
	c.channel = nil
	return err
}

func uniqueConsumerTag(queueName, routingKey string) (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	s := base64.RawURLEncoding.EncodeToString(b)
	return fmt.Sprintf("ctag-%s-%s-%s", queueName, routingKey, s), nil
}
