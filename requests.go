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
)

const azureServiceBusAPIVersion = "2016-07"

func NewRequestUrl(cnx *connectionString, path string) (*url.URL, error) {
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

func NewRequest(cnx *connectionString, url *url.URL, method string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", makeAuthorizationHeader(cnx))

	return req, nil
}

func AddProperty(req *http.Request, key string, value string) {
	req.Header.Set(key, value)
}

func Execute(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
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