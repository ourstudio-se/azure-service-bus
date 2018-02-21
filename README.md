# An Azure Service Bus client implemented in Go

A custom implementation for Azure Service Bus, supporting queues and/or pubsub, and custom message properties.

## Installation

    go get github.com/ourstudio-se/azure-service-bus

## Usage

    client, err := azureservicebus.NewPubSubClient(connectionString, topic, subscription)
    if err != nil {
        log.Fatal(err.Error())
    }

    msg, err := client.PeekLockMessage(30)

The `connectionString` parameter for `NewPubSubClient` and `NewQueueClient` should be in the standard Azure connection string format;

    Endpoint=sb://my-namespace.servicebus.windows.net/;SharedAccessKeyName=MyAccessKeyName;SharedAccessKey=MyAccessKeySecret

See the [examples](https://github.com/ourstudio-se/azure-service-bus/blob/master/examples/) for a full usage example.