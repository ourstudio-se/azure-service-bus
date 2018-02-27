package azureservicebus

import (
	"errors"
	"fmt"
	"strings"

	"net/url"
)

const azureServiceBusFormat = "https://%s.servicebus.windows.net:443"

type connectionString struct {
	url       *url.URL
	namespace string
	keyName   string
	accessKey string
}

func defaultError() error {
	return errors.New("Azure Service Bus ConnectionString was not in a correct format")
}

func namespaceFromURL(urlAsString string) (string, error) {
	url, err := url.Parse(urlAsString)
	if err != nil || !url.IsAbs() || url.Hostname() == "" {
		return "", defaultError()
	}

	return strings.Split(url.Hostname(), ".")[0], nil
}

func extract(kvp string, keyName string) (string, error) {
	parts := strings.SplitN(kvp, "=", 2)

	if len(parts) != 2 || parts[0] != keyName {
		return "", defaultError()
	}

	return parts[1], nil
}

// ParseConnectionString handles standard Azure Service Bus connection
// string formatted strings, and creates a generic instance of a `connectionString`
// from it.
func ParseConnectionString(cnxString string) (*connectionString, error) {
	parts := strings.Split(cnxString, ";")

	if len(parts) < 3 {
		return nil, defaultError()
	}

	urlAsString, err := extract(parts[0], "Endpoint")
	if err != nil {
		return nil, defaultError()
	}

	ns, err := namespaceFromURL(urlAsString)
	if err != nil {
		return nil, defaultError()
	}

	kn, err := extract(parts[1], "SharedAccessKeyName")
	if err != nil {
		return nil, defaultError()
	}

	ak, err := extract(parts[2], "SharedAccessKey")
	if err != nil {
		return nil, defaultError()
	}

	target, err := url.Parse(fmt.Sprintf(azureServiceBusFormat, ns))
	if err != nil {
		return nil, defaultError()
	}

	return &connectionString{target, ns, kn, ak}, nil
}
