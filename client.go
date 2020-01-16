package azureservicebus

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// Client contain methods to communicate with Azure Service Bus over HTTPS
type Client interface {
	Send(message *Message) error
	PeekLockMessage(timeout int) (*Message, error)
	Unlock(message *Message) error
	RenewLock(message *Message) error
	DestructiveRead(timeout int) (*Message, error)
	DeleteMessage(message *Message) error
}

type queueClient struct {
	queueName string
	client    *HTTPRequestClient
}

type pubsubClient struct {
	topic        string
	subscription string
	client       *HTTPRequestClient
}

func send(client *HTTPRequestClient, path string, message *Message) error {
	target, err := client.NewRequestURL(path)
	if err != nil {
		return err
	}

	req, err := client.NewRequest(target, "POST", message.Body)
	if err != nil {
		return err
	}

	for key, value := range message.Properties {
		req.Header[key] = []string{value}
	}

	resp, err := client.Execute(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return nil
	}

	return fmt.Errorf("Could not send message. Server returned error %d", resp.StatusCode)
}

func peekLockMessage(client *HTTPRequestClient, path string, timeout int) (*Message, error) {
	target, err := client.NewRequestURL(path)
	if err != nil {
		return nil, err
	}

	req, err := client.NewRequest(target, "POST", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.ExecuteWithTimeout(req, time.Duration(timeout)*time.Second)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	msg, err := ResponseToMessage(resp)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func unlockMessage(client *HTTPRequestClient, message *Message) error {
	target, err := url.Parse(message.Location)
	if err != nil {
		return err
	}

	req, err := client.NewRequest(target, "PUT", nil)
	if err != nil {
		return err
	}

	resp, err := client.Execute(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("Could not unlock message. Server returned error %d", resp.StatusCode)
}

func renewMessageLock(client *HTTPRequestClient, message *Message) error {
	target, err := url.Parse(message.Location)
	if err != nil {
		return err
	}

	req, err := client.NewRequest(target, "POST", nil)
	if err != nil {
		return err
	}

	resp, err := client.Execute(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("Could not renew message lock. Server returned error %d", resp.StatusCode)
}

func destructiveReadMessage(client *HTTPRequestClient, path string, timeout int) (*Message, error) {
	target, err := client.NewRequestURL(path)
	if err != nil {
		return nil, err
	}

	req, err := client.NewRequest(target, "DELETE", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.ExecuteWithTimeout(req, time.Duration(timeout)*time.Second)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	msg, err := ResponseToMessage(resp)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func deleteMessage(client *HTTPRequestClient, message *Message) error {
	target, err := url.Parse(message.Location)
	if err != nil {
		return err
	}

	req, err := client.NewRequest(target, "DELETE", nil)
	if err != nil {
		return err
	}

	resp, err := client.Execute(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("Could not delete message. Server returned error %d", resp.StatusCode)
}

// Send a new message to the Azure Service Bus Queue
func (c *queueClient) Send(message *Message) error {
	path := fmt.Sprintf("/%s/messages/", c.queueName)
	return send(c.client, path, message)
}

// PeekLockMessage listens for a message without removing it
// from the queue. The timeout should be specified in seconds.
func (c *queueClient) PeekLockMessage(timeout int) (*Message, error) {
	path := fmt.Sprintf("/%s/messages/head?timeout=%d", c.queueName, timeout)
	return peekLockMessage(c.client, path, timeout)
}

// Unlock a message in the queue to enable re-processing
func (c *queueClient) Unlock(message *Message) error {
	return unlockMessage(c.client, message)
}

// RenewLock a message in the queue to keep blocking re-processing
func (c *queueClient) RenewLock(message *Message) error {
	return renewMessageLock(c.client, message)
}

// DestructiveRead a message, removing it from the queue. The timeout
// should be specified in seconds.
func (c *queueClient) DestructiveRead(timeout int) (*Message, error) {
	path := fmt.Sprintf("/%s/messages/head?timeout=%d", c.queueName, timeout)
	return destructiveReadMessage(c.client, path, timeout)
}

// DeleteMessage from the queue
func (c *queueClient) DeleteMessage(message *Message) error {
	return deleteMessage(c.client, message)
}

// Send a new message to the Azure Service Bus publisher
func (c *pubsubClient) Send(message *Message) error {
	path := fmt.Sprintf("/%s/messages/", c.topic)
	return send(c.client, path, message)
}

// PeekLockMessage listens for a message without removing it
// from the subscriber. The timeout should be specified in seconds.
func (c *pubsubClient) PeekLockMessage(timeout int) (*Message, error) {
	path := fmt.Sprintf("/%s/subscriptions/%s/messages/head?timeout=%d", c.topic, c.subscription, timeout)
	return peekLockMessage(c.client, path, timeout)
}

// Unlock a message in the subscription to enable re-processing
func (c *pubsubClient) Unlock(message *Message) error {
	return unlockMessage(c.client, message)
}

// RenewLock a message in the subscription to keep blocking re-processing
func (c *pubsubClient) RenewLock(message *Message) error {
	return renewMessageLock(c.client, message)
}

// DestructiveRead a message, removing it from the subscription. The timeout
// should be specified in seconds.
func (c *pubsubClient) DestructiveRead(timeout int) (*Message, error) {
	path := fmt.Sprintf("/%s/subscriptions/%s/messages/head?timeout=%d", c.topic, c.subscription, timeout)
	return destructiveReadMessage(c.client, path, timeout)
}

// DeleteMessage from the subscription
func (c *pubsubClient) DeleteMessage(message *Message) error {
	return deleteMessage(c.client, message)
}

// NewQueueClient creates a new instance of an Azure Service Bus
// client aimed at queue communication
func NewQueueClient(cnxString string, queueName string) (Client, error) {
	cnx, err := ParseConnectionString(cnxString)
	if err != nil {
		return nil, err
	}

	return &queueClient{
		queueName: queueName,
		client:    NewHTTPRequestClient(cnx),
	}, nil
}

// NewPubSubClient creates a new instance of an Azure Service Bus
// client aimed at either sending messages to a topic or receiving
// messages from a subscription
func NewPubSubClient(cnxString string, topic string, subscription string) (Client, error) {
	cnx, err := ParseConnectionString(cnxString)
	if err != nil {
		return nil, err
	}

	return &pubsubClient{
		topic:        topic,
		subscription: subscription,
		client:       NewHTTPRequestClient(cnx),
	}, nil
}
