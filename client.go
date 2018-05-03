package azureservicebus

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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
	queueName        string
	connectionString *connectionString
}

type pubsubClient struct {
	topic            string
	subscription     string
	connectionString *connectionString
}

func send(cnx *connectionString, path string, message *Message) error {
	target, err := NewRequestURL(cnx, path)
	if err != nil {
		return err
	}

	req, err := NewRequest(cnx, target, "POST", message.Body)
	if err != nil {
		return err
	}

	for key, value := range message.Properties {
		req.Header[key] = []string{value}
	}

	resp, err := Execute(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return err
	}

	statusCode := resp.StatusCode
	if statusCode == http.StatusOK || statusCode == http.StatusCreated {
		return nil
	}

	return fmt.Errorf("Could not send message. Server returned error %d", statusCode)
}

func peekLockMessage(cnx *connectionString, path string, timeout int) (*Message, error) {
	target, err := NewRequestURL(cnx, path)
	if err != nil {
		return nil, err
	}

	req, err := NewRequest(cnx, target, "POST", nil)
	if err != nil {
		return nil, err
	}

	resp, err := Execute(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	msg, err := ResponseToMessage(resp)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func unlockMessage(cnx *connectionString, message *Message) error {
	target, err := url.Parse(message.Location)
	if err != nil {
		return err
	}

	req, err := NewRequest(cnx, target, "PUT", nil)
	if err != nil {
		return err
	}

	resp, err := Execute(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return err
	}

	statusCode := resp.StatusCode
	if statusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("Could not unlock message. Server returned error %d", statusCode)
}

func renewMessageLock(cnx *connectionString, message *Message) error {
	target, err := url.Parse(message.Location)
	if err != nil {
		return err
	}

	req, err := NewRequest(cnx, target, "POST", nil)
	if err != nil {
		return err
	}

	resp, err := Execute(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return err
	}

	statusCode := resp.StatusCode
	if statusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("Could not renew message lock. Server returned error %d", statusCode)
}

func destructiveReadMessage(cnx *connectionString, path string, timeout int) (*Message, error) {
	target, err := NewRequestURL(cnx, path)
	if err != nil {
		return nil, err
	}

	req, err := NewRequest(cnx, target, "DELETE", nil)
	if err != nil {
		return nil, err
	}

	resp, err := Execute(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	msg, err := ResponseToMessage(resp)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func deleteMessage(cnx *connectionString, message *Message) error {
	target, err := url.Parse(message.Location)
	if err != nil {
		return err
	}

	req, err := NewRequest(cnx, target, "DELETE", nil)
	if err != nil {
		return err
	}

	resp, err := Execute(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	statusCode := resp.StatusCode
	if statusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("Could not delete message. Server returned error %d", statusCode)
}

func (c *queueClient) Send(message *Message) error {
	path := fmt.Sprintf("/%s/messages/", c.queueName)
	return send(c.connectionString, path, message)
}

func (c *queueClient) PeekLockMessage(timeout int) (*Message, error) {
	path := fmt.Sprintf("/%s/messages/head?timeout=%d", c.queueName, timeout)
	return peekLockMessage(c.connectionString, path, timeout)
}

func (c *queueClient) Unlock(message *Message) error {
	return unlockMessage(c.connectionString, message)
}

func (c *queueClient) RenewLock(message *Message) error {
	return renewMessageLock(c.connectionString, message)
}

func (c *queueClient) DestructiveRead(timeout int) (*Message, error) {
	path := fmt.Sprintf("/%s/messages/head?timeout=%d", c.queueName, timeout)
	return destructiveReadMessage(c.connectionString, path, timeout)
}

func (c *queueClient) DeleteMessage(message *Message) error {
	return deleteMessage(c.connectionString, message)
}

func (c *pubsubClient) Send(message *Message) error {
	path := fmt.Sprintf("/%s/messages/", c.topic)
	return send(c.connectionString, path, message)
}

func (c *pubsubClient) PeekLockMessage(timeout int) (*Message, error) {
	path := fmt.Sprintf("/%s/subscriptions/%s/messages/head?timeout=%d", c.topic, c.subscription, timeout)
	return peekLockMessage(c.connectionString, path, timeout)
}

func (c *pubsubClient) Unlock(message *Message) error {
	return unlockMessage(c.connectionString, message)
}

func (c *pubsubClient) RenewLock(message *Message) error {
	return renewMessageLock(c.connectionString, message)
}

func (c *pubsubClient) DestructiveRead(timeout int) (*Message, error) {
	path := fmt.Sprintf("/%s/subscriptions/%s/messages/head?timeout=%d", c.topic, c.subscription, timeout)
	return destructiveReadMessage(c.connectionString, path, timeout)
}

func (c *pubsubClient) DeleteMessage(message *Message) error {
	return deleteMessage(c.connectionString, message)
}

// NewQueueClient creates a new instance of an Azure Service Bus
// client aimed at queue communication
func NewQueueClient(cnxString string, queueName string) (Client, error) {
	var c Client
	var cnx *connectionString
	cnx, err := ParseConnectionString(cnxString)
	if err != nil {
		return nil, err
	}

	c = &queueClient{
		queueName:        queueName,
		connectionString: cnx,
	}
	return c, nil
}

// NewPubSubClient creates a new instance of an Azure Service Bus
// client aimed at either sending messages to a topic or receiving
// messages from a subscription
func NewPubSubClient(cnxString string, topic string, subscription string) (Client, error) {
	var c Client
	var cnx *connectionString
	cnx, err := ParseConnectionString(cnxString)
	if err != nil {
		return nil, err
	}

	c = &pubsubClient{
		topic:            topic,
		subscription:     subscription,
		connectionString: cnx,
	}
	return c, nil
}
