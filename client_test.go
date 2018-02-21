package azureservicebus

import "testing"

func TestNewQueueClient(t *testing.T) {
	client, err := NewQueueClient("Endpoint=sb://test.servicebus.windows.net/;SharedAccessKeyName=TestSharedAccessKey;SharedAccessKey=TestSharedAccessKey", "test-queue")
	if err != nil {
		t.Errorf("Could not create queue client.")
	}
	if client == nil {
		t.Errorf("Queue client is nil.")
	}
}

func TestNewPubSubClient(t *testing.T) {
	client, err := NewPubSubClient("Endpoint=sb://test.servicebus.windows.net/;SharedAccessKeyName=TestSharedAccessKey;SharedAccessKey=TestSharedAccessKey", "test-topic", "test-subscription")
	if err != nil {
		t.Errorf("Could not create pubsub client.")
	}
	if client == nil {
		t.Errorf("Pubsub client is nil.")
	}
}
