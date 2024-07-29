package networking

import (
	"bytes"
	"net/http"
	"strings"
	"time"
)

// SendSoap send soap message
func SendSoap(httpClient *http.Client, endpoint, message string, username, password string, urlType string) (*http.Response, error) {
	if urlType == "https" {
		endpoint = strings.Replace(endpoint, "http://", "https://", -1)
	}

	t := NewTransport(username, password, urlType)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(message))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")

	resp, err := t.RoundTrip(req)
	if err != nil {
		return nil,err
	}
	
	//resp, err := httpClient.Post(endpoint, "application/soap+xml; charset=utf-8", bytes.NewBufferString(message))
	//if err != nil {
	//	return resp, err
	//}

	return resp, nil
}

//// SendSoap send soap message
func SendSoapEx(endpoint, message string, username, password string, urlType string, encryptionType string) RestResponse {
	if urlType == "https" {
		endpoint = strings.Replace(endpoint, "http://", "https://", -1)
	}

	headers := make(map[string]string)
	headers["Content-Type"] = "application/soap+xml; charset=utf-8"
	cli := RestClientAuth{}
	cli.SetAuth(username, password)
	return cli.Request("POST", endpoint, []byte(message), headers, time.Second * 5, encryptionType)
}
