package pubsub

import (
	"strings"
	"sync"

	"github.com/parashmaity/fleare/config"
	"github.com/parashmaity/fleare/internal/comm"
	"github.com/parashmaity/fleare/internal/logger"
	"github.com/parashmaity/fleare/internal/utils"
	"github.com/parashmaity/fleare/server/eventloop"
)

// Message represents a message to be published
type Message struct {
	Topic   string
	Payload interface{}
}

// Subscriber represents a subscriber
type Subscriber struct {
	ID      string
	Channel chan Message
	Conn    *eventloop.Conn
}

// Broker manages subscriptions and message distribution
type Broker struct {
	subscribers map[string]map[string]*Subscriber // topic -> clientID -> Subscriber
	mutex       sync.RWMutex
}

// NewBroker creates a new Broker instance
func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[string]map[string]*Subscriber),
	}
}

// HandlePubSub handles client pubsub commands
func (b *Broker) HandlePubSub(conn *eventloop.Conn, clientID string, c *comm.Command) {
	switch strings.ToLower(c.Command) {
	case "subscribe":
		if len(c.Args) < 1 {
			b.replyError(conn, clientID, "missing topic for subscribe")
			return
		}
		topic := c.Args[0]
		b.Subscribe(topic, clientID, conn)

	case "unsubscribe":
		if len(c.Args) < 1 {
			b.replyError(conn, clientID, "missing topic for unsubscribe")
			return
		}
		topic := c.Args[0]
		b.Unsubscribe(topic, clientID)

	case "publish":
		if len(c.Args) < 2 {
			b.replyError(conn, clientID, "missing topic or message for publish")
			return
		}
		topic := c.Args[0]
		payload := c.Args[1]
		b.Publish(conn, topic, utils.EnsureUnmarshal(payload))

		conn.WriteSync(&comm.Response{
			ClientId: clientID,
			Result:   []byte(""),
			Status:   config.STATUS_SUCCESS,
		})

	default:
		b.replyError(conn, clientID, "unknown pubsub command")
	}
}

// Subscribe adds a new subscriber to a topic
func (b *Broker) Subscribe(topic, clientID string, conn *eventloop.Conn) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, ok := b.subscribers[topic]; !ok {
		b.subscribers[topic] = make(map[string]*Subscriber)
	}

	// avoid duplicate subscriptions
	if _, exists := b.subscribers[topic][clientID]; exists {
		return
	}

	sub := &Subscriber{
		ID:      clientID,
		Conn:    conn,
		Channel: make(chan Message, 100), // buffered channel
	}

	b.subscribers[topic][clientID] = sub

	// start a goroutine to deliver messages to this client
	go b.deliverMessages(sub, topic)

	logger.Info("Subscribe client", map[string]any{
		"subscribe": topic,
		"clientID":  clientID,
	})
}

// Unsubscribe removes a subscriber from a topic
func (b *Broker) Unsubscribe(topic, clientID string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	logger.Info("Unsubscribe client", map[string]any{
		"subscribe": topic,
		"clientID":  clientID,
	})

	if subs, ok := b.subscribers[topic]; ok {
		if sub, exists := subs[clientID]; exists {
			close(sub.Channel)
			delete(subs, clientID)
		}
		if len(subs) == 0 {
			delete(b.subscribers, topic)
		}
	}
}

// Publish sends a message to all subscribers of a topic
func (b *Broker) Publish(conn *eventloop.Conn, topic string, payload interface{}) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	logger.Info("Published client", map[string]any{
		"subscribe-topic": topic,
	})

	msg := Message{Topic: topic, Payload: payload}
	if subs, ok := b.subscribers[topic]; ok {
		for _, sub := range subs {
			select {
			case sub.Channel <- msg:
			default:
				// Log instead of printing to stdout for better production logging
				logger.Warn("Client channel full, dropping message", map[string]any{
					"clientID": sub.ID,
					"topic":    topic,
				})
			}
		}
	}
}

// deliverMessages continuously sends messages from channel to client
// MODIFIED: This function now handles write errors to detect disconnections.
func (b *Broker) deliverMessages(sub *Subscriber, topic string) {

	for msg := range sub.Channel {
		v := utils.ObjectToByte(msg.Payload)
		err := sub.Conn.WriteSync(&comm.Response{
			ClientId: sub.ID,
			Result:   []byte(v),
			Status:   config.STATUS_SUCCESS,
			Topic:    topic,
		})

		// If WriteSync returns an error, the client has likely disconnected.
		if err != nil {
			logger.Debug("Client disconnected, cleaning up subscriptions.", map[string]any{
				"clientID": sub.ID,
				"error":    err.Error(),
			})
			// Clean up all subscriptions for this client and exit the goroutine.
			b.HandleClientDisconnect(sub.ID)
			return
		}
	}

}

// HandleClientDisconnect removes a client from all topics they are subscribed to.
// NEW: This method is the central cleanup point for a disconnected client.
func (b *Broker) HandleClientDisconnect(clientID string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Iterate over all topics to find and remove the client
	for topic, subs := range b.subscribers {
		if sub, exists := subs[clientID]; exists {
			// Close the channel to unblock the deliverMessages goroutine and allow it to exit
			close(sub.Channel)
			// Remove the subscriber from the map
			delete(subs, clientID)
			logger.Debug("Unsubscribed client from topic due to disconnect", map[string]any{
				"clientID": clientID,
				"topic":    topic,
			})
		}
		// If the topic has no more subscribers, remove the topic itself
		if len(subs) == 0 {
			delete(b.subscribers, topic)
		}
	}
}

// Close cleans up all resources
func (b *Broker) Close() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for _, subs := range b.subscribers {
		for _, sub := range subs {
			close(sub.Channel)
		}
	}
	b.subscribers = make(map[string]map[string]*Subscriber)
}

func (b *Broker) replyError(conn *eventloop.Conn, clientID, msg string) {
	conn.WriteSync(&comm.Response{
		ClientId: clientID,
		Result:   []byte(msg),
		Status:   config.STATUS_ERROR,
	})
}
