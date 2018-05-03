package azureservicebus

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Message maps to a Azure Service Bus message with broker
// properties and properties for custom properties as well
type Message struct {
	MessageID              string `json:"MessageId"`
	DeliveryCount          int
	EnqueuedSequenceNumber int
	EnqueuedTimeUtc        dateTime
	LockToken              string
	LockedUntilUtc         dateTime
	PartitionKey           string
	SequenceNumber         int
	State                  string
	TimeToLive             float64

	Location string

	Properties map[string]string `json:"Properties"`
	Body       []byte
}

type dateTime struct {
	time.Time
}

func (t *dateTime) UnmarshalJSON(b []byte) (err error) {
	dt := strings.Trim(string(b), "\"")
	if dt == "null" || dt == "" {
		t.Time = time.Time{}
		return
	}
	t.Time, err = time.Parse(time.RFC1123, dt)
	return
}

// ResponseToMessage reads a response byte stream and
// creates a new Message instance from it
func ResponseToMessage(resp *http.Response) (*Message, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	props := resp.Header.Get("brokerproperties")
	location := resp.Header.Get("location")

	var message Message
	if err := json.Unmarshal([]byte(props), &message); err != nil {
		return nil, err
	}

	message.Location = location
	message.Body = body

	properties := make(map[string]string)
	presets := map[string]int{
		"brokerproperties":          1,
		"strict-transport-security": 1,
		"content-type":              1,
		"location":                  1,
		"server":                    1,
		"date":                      1,
	}
	for key, value := range resp.Header {
		if presets[strings.ToLower(key)] != 1 {
			properties[strings.ToLower(key)] = strings.Trim(value[0], "\n\r\t\"'")
		}
	}

	if len(properties) > 0 {
		message.Properties = properties
	}
	if err != nil {
		return &message, err
	}

	return &message, nil
}
