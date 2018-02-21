package azureservicebus

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Message struct {
	MessageID              string `json:"MessageId"`
	DeliveryCount          int
	EnqueuedSequenceNumber int
	EnqueuedTimeUtc        Time
	LockToken              string
	LockedUntilUtc         Time
	PartitionKey           string
	SequenceNumber         int
	State                  string
	TimeToLive             float64

	Location string

	CustomProperties map[string]string

	Body []byte
}

type Time struct {
	time.Time
}

func (t *Time) UnmarshalJSON(b []byte) (err error) {
	dateTime := strings.Trim(string(b), "\"")
	if dateTime == "null" || dateTime == "" {
		t.Time = time.Time{}
		return
	}
	t.Time, err = time.Parse(time.RFC1123, dateTime)
	return
}

func ResponseToMessage(resp *http.Response, propertyHeaders []string) (*Message, error) {
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	props := resp.Header.Get("BrokerProperties")
	location := resp.Header.Get("Location")

	var message Message
	if err := json.Unmarshal([]byte(props), &message); err != nil {
		return nil, err
	}

	message.Location = location
	message.Body = body

	message.CustomProperties, err = extractProperties(resp, propertyHeaders)
	if err != nil {
		return &message, err
	}

	return &message, nil
}

func extractProperties(resp *http.Response, propertyHeaders []string) (map[string]string, error) {
	var customProperties map[string]string
	customProperties = make(map[string]string)

	makeHeader, err := converter()
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(propertyHeaders); i++ {
		headerName := makeHeader(propertyHeaders[i])
		headerValue := strings.Trim(resp.Header.Get(headerName), "\n\r\t\"")

		if headerValue != "" {
			customProperties[propertyHeaders[i]] = headerValue
		}
	}

	return customProperties, nil
}

func converter() (func(string) string, error) {
	reg, err := regexp.Compile("[^a-z0-9]")
	if err != nil {
		return nil, err
	}

	return func(headerName string) string {
		return reg.ReplaceAllString(strings.ToLower(headerName), "")
	}, nil
}
