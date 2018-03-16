package azureservicebus

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestNewRequestURL(t *testing.T) {
	hostname := "test.servicebus.windows.net"
	path := "/test"

	cnx, err := ParseConnectionString(fmt.Sprintf("Endpoint=sb://%s/;SharedAccessKeyName=TestSharedAccessKey;SharedAccessKey=TestSharedAccessKey", hostname))
	if err != nil {
		t.Errorf("Connectionstring could not be parsed.")
	}

	url, err := NewRequestURL(cnx, path)
	if err != nil {
		t.Errorf("Could not create request URL.")
	}

	if url == nil {
		t.Errorf("Request URL is null.")
	}
	if url.Scheme != "https" {
		t.Errorf("Request URL is not using SSL/TLS.")
	}
	if url.Hostname() != hostname {
		t.Errorf("Request URL did not use correct hostname for Azure Service Bus.")
	}
	if url.Path != path {
		t.Errorf("Request URL did not use specified path.")
	}
	if url.Query().Get("api-version") == "" {
		t.Errorf("Request URL does not contain Azure Service Bus API version.")
	}
}

func TestNewRequestWithEmptyBody(t *testing.T) {
	cnx, err := ParseConnectionString("Endpoint=sb://test.servicebus.windows.net/;SharedAccessKeyName=TestSharedAccessKey;SharedAccessKey=TestSharedAccessKey")
	if err != nil {
		t.Errorf("Connectionstring could not be parsed.")
	}

	url, err := NewRequestURL(cnx, "/test")
	if err != nil {
		t.Errorf("Could not create request URL.")
	}

	req, err := NewRequest(cnx, url, "POST", nil)
	if err != nil {
		t.Errorf("Could not create new request.")
	}

	if req.Header.Get("Accept") != "application/json" {
		t.Errorf("Request did not use correct Accept header.")
	}
	if req.Header.Get("Authorization") == "" {
		t.Errorf("Request did not use correct Authorization header.")
	}
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		t.Errorf("Could not read request body.")
	}
	if len(data) != 0 {
		t.Errorf("Request did contain body when none was specified.")
	}
}

func TestNewRequestWithBody(t *testing.T) {
	cnx, err := ParseConnectionString("Endpoint=sb://test.servicebus.windows.net/;SharedAccessKeyName=TestSharedAccessKey;SharedAccessKey=TestSharedAccessKey")
	if err != nil {
		t.Errorf("Connectionstring could not be parsed.")
	}

	url, err := NewRequestURL(cnx, "/test")
	if err != nil {
		t.Errorf("Could not create request URL.")
	}

	body := []byte("test-body")
	req, err := NewRequest(cnx, url, "POST", body)
	if err != nil {
		t.Errorf("Could not create new request.")
	}

	if req.Header.Get("Accept") != "application/json" {
		t.Errorf("Request did not use correct Accept header.")
	}
	if req.Header.Get("Authorization") == "" {
		t.Errorf("Request did not use correct Authorization header.")
	}
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		t.Errorf("Could not read request body.")
	}
	if len(data) != len(body) {
		t.Errorf("Request did not contain correct body data.")
	}
}

func TestAddProperty(t *testing.T) {
	cnx, err := ParseConnectionString("Endpoint=sb://test.servicebus.windows.net/;SharedAccessKeyName=TestSharedAccessKey;SharedAccessKey=TestSharedAccessKey")
	if err != nil {
		t.Errorf("Connectionstring could not be parsed.")
	}

	url, err := NewRequestURL(cnx, "/test")
	if err != nil {
		t.Errorf("Could not create request URL.")
	}

	req, err := NewRequest(cnx, url, "POST", nil)
	if err != nil {
		t.Errorf("Could not create new request.")
	}

	AddProperty(req, "Content-Encoding", "gzip")
	AddProperty(req, "Content-Type", "application/json")

	if req.Header.Get("Content-Encoding") != "gzip" {
		t.Errorf("Request did not contain specified custom property Content-Encoding.")
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Request did not contain specified custom property Content-Type.")
	}
}
