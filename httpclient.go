package azureservicebus

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gojektech/heimdall"
	"github.com/gojektech/heimdall/httpclient"
)

const azureServiceBusAPIVersion = "2016-07"

type HTTPRequestClient struct {
	client           *httpclient.Client
	connectionString *connectionString
}

func NewHTTPRequestClient(cnx *connectionString) *HTTPRequestClient {
	initalTimeout := 5 * time.Second
	maxTimeout := 10 * time.Second
	exponentFactor := 2.0
	maximumJitterInterval := 100 * time.Millisecond

	backoff := heimdall.NewExponentialBackoff(initalTimeout, maxTimeout, exponentFactor, maximumJitterInterval)
	retrier := heimdall.NewRetrier(backoff)

	timeout := 5 * time.Second
	client := httpclient.NewClient(
		httpclient.WithHTTPTimeout(timeout),
		httpclient.WithRetrier(retrier),
		httpclient.WithRetryCount(4),
	)

	return &HTTPRequestClient{
		client:           client,
		connectionString: cnx,
	}
}

// NewRequestURL creates an Azure Service Bus URL (with versioning)
// from a the specified connection string and action path
func (hrc *HTTPRequestClient) NewRequestURL(path string) (*url.URL, error) {
	query := hrc.connectionString.url.Query()
	query.Set("api-version", azureServiceBusAPIVersion)

	target := fmt.Sprintf("%s%s", hrc.connectionString.url.String(), path)
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	u.RawQuery = query.Encode()
	return u, nil
}

// NewRequest creates a new http.Request instance with the correct
// headers set for communication with an Azure Service Bus
func (hrc *HTTPRequestClient) NewRequest(url *url.URL, method string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", makeAuthorizationHeader(hrc.connectionString))

	return req, nil
}

// Execute is an abstraction for actually making a HTTP request
// to the Azure Service Bus, implemented with Pester to support
// retry and back off functionality
func (hrc *HTTPRequestClient) Execute(req *http.Request) (*http.Response, error) {
	return hrc.client.Do(req)
}

func makeAuthorizationHeader(cnx *connectionString) string {
	ticks := time.Now().Add(300 * time.Second).Round(time.Second).Unix()
	expires := strconv.Itoa(int(ticks))

	uri := url.QueryEscape(cnx.url.String())

	hash := hmac.New(sha256.New, []byte(cnx.accessKey))
	hash.Write([]byte(uri + "\n" + expires))
	signature := url.QueryEscape(base64.StdEncoding.EncodeToString(hash.Sum(nil)))

	return fmt.Sprintf("SharedAccessSignature sig=%s&se=%s&skn=%s&sr=%s", signature, expires, cnx.keyName, uri)
}
