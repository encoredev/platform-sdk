package encorecloud

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog"
	"go.encore.dev/platform-sdk/encorecloud/types"
	"go.encore.dev/platform-sdk/internal/jsonerr"
	"go.encore.dev/platform-sdk/pkg/auth"
)

const (
	KeepAliveInterval = 5 * time.Second
)

// PublishToTopic publishes the specified attrs and data to the topic specified by topicID.
//
// It returns the message ID of the published message from the underlying message broker and
// any error encountered.
func (c *Client) PublishToTopic(ctx context.Context, topicID string, attrs map[string]string, data []byte) (msgID string, err error) {
	params := &types.PublishParams{
		Attributes: attrs,
		Payload:    data,
	}
	resp := &types.PublishResponse{}

	err = c.client.SignedPost(
		ctx,
		fmt.Sprintf("/v1/pubsub/%s/publish", url.PathEscape(topicID)),
		auth.PubsubMsg, auth.Create,
		params, resp,

		[]byte(topicID),
	)
	if err != nil {
		return "", fmt.Errorf("unable to sign publish request: %w", err)
	}

	return resp.MessageID, nil
}

// CreateSubscriptionHandler returns a [http.HandlerFunc] that can be used to handle
// subscription push requests from EncoreCloud.
//
// Encore Cloud will send a POST request to the endpoint with a JSON encoded [pushPayload] as the body.
// The request will be signed with the latest Encore Cloud auth key for this application.
//
// Once the request is received and verified, the user's subscription function will be called with the decoded
// payload, while simultaneously an event stream will be sent back to Encore Cloud to indicate that the request
// is being processed, with keepalive messages being sent every 5 seconds.
//
// If the subscription function returns an error, the event stream will be closed with the error message.
// If the subscription function returns successfully, the event stream will be closed with a success message.
//
// The Encore Cloud server will wait for a valid end response from the event stream before closing the connection and
// acknowledging the message with the underlying message broker.
//
// If the event stream is closed without a valid end response, the message will be nacked and retried by Encore Cloud.
//
// If the request is closed by Encore Cloud while a subscription function is still running, the context of the function
// will be cancelled, as this means Encore Cloud has failed to receive a keepalive message from the event stream and has
// assumed the request has failed.
//
// The events on the stream will be one of these types:
// - "keepalive" - A message to inform the server that the client is still processing.
// - "ack" - A message to confirm the client has successfully processed the message.
// - "nack" - A message to tell the server the client failed to process the message and it should be retried.
func (c *Client) CreateSubscriptionHandler(subscriptionID string, logger *zerolog.Logger, callback types.SubscriptionCallback) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Decode the request
		payload := &types.SubscriptionPushParams{}
		err := c.client.VerifyAndDecodeRequest(
			req,
			auth.PubsubMsg, auth.Read,
			payload,
			[]byte(subscriptionID),
		)
		if err != nil {
			logger.Err(err).Msg("error while verifying PubSub subscription message")
			jsonerr.Error(w, err, http.StatusUnauthorized)
			return
		}

		// Ensure we can flush the responses
		flusher, ok := w.(http.Flusher)
		if !ok {
			err = errors.New("unable to cast http.ResponseWriter to http.Flusher")
			logger.Err(err).Msg("error while setting up flushing response")
			jsonerr.Error(w, err, http.StatusInternalServerError)
			return
		}

		// Start the event stream
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		// Run the subscription function in a goroutine
		response := make(chan error, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic while processing PubSub message: %v", r)
					response <- err
				}
				close(response)
			}()

			response <- callback(
				req.Context(),
				payload.MessageID, payload.PublishTime, payload.DeliveryAttempt,
				payload.Attributes, payload.Data,
			)
		}()

		// Wait for the function to complete or the request to be cancelled
		var firstError error
		var finished bool
		keepAliveTimeout := time.NewTicker(KeepAliveInterval)
		defer keepAliveTimeout.Stop()

		for !finished {
			select {
			case <-req.Context().Done():
				logger.Err(err).Msg("PubSub push endpoint closed by Encore Cloud before subscription function completed")
				return

			case <-keepAliveTimeout.C:
				// Send a keepalive message
				if _, err := fmt.Fprintf(w, "event: keepalive\ndata: \n\n"); err != nil {
					logger.Err(err).Msg("error while sending keepalive message")
				}
				flusher.Flush()

			case err, done := <-response:
				if done {
					finished = true
				} else if firstError == nil {
					firstError = err
				}
			}
		}

		// Now that the subscription function has completed, send the end message
		if firstError != nil {
			logger.Err(firstError).Msg("error while handling PubSub subscription message")

			if _, err := fmt.Fprintf(w, "event: nack\ndata: %s\n\n", firstError.Error()); err != nil {
				logger.Err(err).Msg("error while sending nack message")
			}
		} else {
			if _, err := fmt.Fprintf(w, "event: ack\ndata: \n\n"); err != nil {
				logger.Err(err).Msg("error while sending ack message")
			}
		}
		flusher.Flush()

		// Now wait for the request to be closed by Encore Cloud (upto 5 seconds)
		select {
		case <-req.Context().Done():
			// If the request is closed by Encore Cloud, the context will be cancelled, this is a sign that it has processed
			// our end message successfully

		case <-time.After(KeepAliveInterval):
			// If we get here, the request was not closed by Encore Cloud, so we should log an error
			// and return
			logger.Err(err).Msg("PubSub push connection was not closed by Encore Cloud after ack/nack message sent")
		}
	}
}
