package main

import (
	"log"
	"time"

	azureservicebus "github.com/ourstudio-se/azure-service-bus"
)

func main() {
	connectionString := "Endpoint=sb://my-namespace.servicebus.windows.net/;SharedAccessKeyName=MyAccessKeyName;SharedAccessKey=MyAccessKeySecret"
	topic := "my-test-topic"
	subscription := "my-test-subscription"

	client, err := azureservicebus.NewPubSubClient(connectionString, topic, subscription)
	if err != nil {
		log.Fatal(err)
	}

	messageText := "My PubSub Message"

	message := &azureservicebus.Message{}
	message.Body = []byte(messageText)

	log.Println("Message sent:")
	log.Println(messageText)

	err = client.Send(message)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(3 * time.Second)

	msg, err := client.PeekLockMessage(30)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Message received:")
	log.Println(string(msg.Body))

	err = client.DeleteMessage(msg)
	if err != nil {
		log.Fatal(err)
	}
}
