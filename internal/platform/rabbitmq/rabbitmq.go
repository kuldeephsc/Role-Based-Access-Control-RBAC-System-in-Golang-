package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EventsExchange = "rbac.events"
	dlxExchange    = "rbac.events.dlx"
)

type Conn struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// Connect opens the connection and declares the full topology described
// in the architecture spec (§5.7): the rbac.events topic exchange, three
// consumer queues with their bindings, and a dead-letter exchange + DLQ
// per queue so a message that fails repeatedly is never silently dropped.
func Connect(url string) (*Conn, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	c := &Conn{conn: conn, ch: ch}
	if err := c.setupTopology(); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

func (c *Conn) Channel() *amqp.Channel { return c.ch }

func (c *Conn) Close() {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *Conn) setupTopology() error {
	if err := c.ch.ExchangeDeclare(EventsExchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if err := c.ch.ExchangeDeclare(dlxExchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	queues := []struct {
		name     string
		bindings []string
	}{
		{"audit.queue", []string{"#"}},
		{"notification.queue", []string{"user.created", "role.assigned"}},
		{"analytics.queue", []string{"login.success", "login.failure", "user.created"}},
	}

	for _, q := range queues {
		dlqRoutingKey := q.name + ".dlq"

		args := amqp.Table{
			"x-dead-letter-exchange":    dlxExchange,
			"x-dead-letter-routing-key": dlqRoutingKey,
		}
		if _, err := c.ch.QueueDeclare(q.name, true, false, false, false, args); err != nil {
			return err
		}
		for _, key := range q.bindings {
			if err := c.ch.QueueBind(q.name, key, EventsExchange, false, nil); err != nil {
				return err
			}
		}

		dlqName := dlqRoutingKey
		if _, err := c.ch.QueueDeclare(dlqName, true, false, false, false, nil); err != nil {
			return err
		}
		if err := c.ch.QueueBind(dlqName, dlqRoutingKey, dlxExchange, false, nil); err != nil {
			return err
		}
	}
	return nil
}
