package types

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// PublishParams is the parameters for publishing a message to a topic.
type PublishParams struct {
	Attributes map[string]string `json:"attributes,omitempty" encore:"sensitive"` // Optional attributes for this message.
	Payload    json.RawMessage   `json:"payload" encore:"sensitive"`              // The message payload.
}

func (p *PublishParams) DeterministicBytes() []byte {
	b, _ := json.Marshal(p)
	return b
}

func (p *PublishParams) Validate() error {
	if len(p.Payload) == 0 {
		return errors.New("payload must be provided")
	}

	return nil
}

// PublishResponse is the response from publishing a message to a topic.
type PublishResponse struct {
	MessageID string `json:"message_id"`
}

// SubscriptionPushParams is the payload that Encore Cloud will generate
// when pushing a subscription attempt to a push endpoint.
type SubscriptionPushParams struct {
	Data            []byte            `json:"data"`
	Attributes      map[string]string `json:"attributes,omitempty"`
	MessageID       string            `json:"message_id"`
	PublishTime     time.Time         `json:"publish_time"`
	DeliveryAttempt int               `json:"delivery_attempt"`
}

func (s *SubscriptionPushParams) DeterministicBytes() []byte {
	b, _ := json.Marshal(s)
	return b
}

// SubscriptionCallback is the callback function that will be invoked when a subscription
// receives a message.
type SubscriptionCallback = func(ctx context.Context, msgID string, publishTime time.Time, deliveryAttempt int, attrs map[string]string, data []byte) error
