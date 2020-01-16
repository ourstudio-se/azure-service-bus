package azureservicebus

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
)

// NewRequestURL creates an Azure Service Bus URL (with versioning)
// from a the specified connection string and action path
//
// [Deprecated]: use HTTPRequestClient instead
func NewRequestURL(cnx *connectionString, path string) (*url.URL, error) {
	baseurl := cnx.url

	query := baseurl.Query()
	query.Set("api-version", azureServiceBusAPIVersion)

	target := fmt.Sprintf("%s%s", baseurl.String(), path)
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	u.RawQuery = query.Encode()
	return u, nil
}

// NewRequest creates a new http.Request instance with the correct
// headers set for communication with an Azure Service Bus
//
// [Deprecated]: use HTTPRequestClient instead
func NewRequest(cnx *connectionString, url *url.URL, method string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", makeAuthorizationHeader(cnx))

	return req, nil
}

// Execute is an abstraction for actually making a HTTP request
// to the Azure Service Bus, implemented with Pester to support
// retry and back off functionality
//
// [Deprecated]: use HTTPRequestClient instead
func Execute(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
