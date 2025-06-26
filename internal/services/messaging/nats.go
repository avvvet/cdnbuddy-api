package messaging

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type NATSClient struct {
	conn *nats.Conn
}

func NewNATSClient(url string) (*NATSClient, error) {
	opts := []nats.Option{
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(10),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("❌ NATS disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("🔄 NATS reconnected to %v", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Printf("🔒 NATS connection closed")
		}),
	}

	conn, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Connected to NATS at %s", url)
	return &NATSClient{conn: conn}, nil
}

func (n *NATSClient) Close() {
	if n.conn != nil {
		n.conn.Close()
	}
}

func (n *NATSClient) Publish(subject string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return n.conn.Publish(subject, payload)
}

func (n *NATSClient) PublishWithReply(subject, reply string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return n.conn.PublishRequest(subject, reply, payload)
}

func (n *NATSClient) Subscribe(subject string, handler func(msg *nats.Msg)) (*nats.Subscription, error) {
	return n.conn.Subscribe(subject, handler)
}

func (n *NATSClient) QueueSubscribe(subject, queue string, handler func(msg *nats.Msg)) (*nats.Subscription, error) {
	return n.conn.QueueSubscribe(subject, queue, handler)
}

func (n *NATSClient) Request(subject string, data interface{}, timeout time.Duration) (*nats.Msg, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return n.conn.Request(subject, payload, timeout)
}

func (n *NATSClient) IsConnected() bool {
	return n.conn != nil && n.conn.IsConnected()
}
