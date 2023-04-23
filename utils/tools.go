package utils

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

//this function constructs http requests using received information
// It constructs an HTTP request with the given information...
// ...and calls ExternalRequestTimer to make the reques
func Request(request string, headers map[string][]string, urlPath string, method string) (string, error) {

	reqURL, _ := url.Parse(urlPath)

	reqBody := io.NopCloser(strings.NewReader(request))
	req := &http.Request{
		Method: method,
		URL:    reqURL,
		Header: headers,
		Body:   reqBody,
	}

	res, err := ExternalRequestTimer(req)
	if err != nil {
		logrus.Errorf("SEND REQUEST | URL : %s | METHOD : %s | BODY : %s | ERROR : %v", urlPath, method, request, err)
		return "", err
	}

	data, _ := io.ReadAll(res.Body)
	defer res.Body.Close()
	resbody := string(data)

	logrus.Infof("SEND REQUEST | URL : %s | METHOD : %s | BODY : %s | STATUS : %s | HTTP_CODE : %d | RESPONSE : %s", urlPath, method, request, res.Status, res.StatusCode, resbody)

	if res.StatusCode > 299 || res.StatusCode <= 199 {
		logrus.Errorf("SEND REQUEST | URL : %s | METHOD : %s | BODY : %s | STATUS : %s | HTTP_CODE : %d", urlPath, method, request, res.Status, res.StatusCode)
		return resbody, fmt.Errorf("%d", res.StatusCode)
	}

	return resbody, nil
}

//This function takes an HTTP request as input and adds timing information to it using an httptrace.ClientTrace object
//It then makes the request using the default HTTP transport with the RoundTrip function and returns the response and any errors that occur.
func ExternalRequestTimer(req *http.Request) (*http.Response, error) {

	var start, connect, dns, tlsHandshake time.Time
// ClientTrace is a set of hooks to run at various stages of an outgoing
// HTTP request. Any particular hook may be nil. Functions may be
// called concurrently from different goroutines and some may be called
// after the request has completed or failed.
//
// ClientTrace currently traces a single HTTP request & response
// during a single round trip and has no hooks that span a series
// of redirected requests.
	trace := &httptrace.ClientTrace{
	// DNSDone is called when a DNS lookup ends.
		DNSStart: func(dsi httptrace.DNSStartInfo) {
			dns = time.Now()
			logrus.Debug(dsi)
		},

	// DNSDone is called when a DNS lookup ends.
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			logrus.Debug(ddi)
			logrus.Infof("DNS Done: %v", time.Since(dns))
		},

// TLSHandshakeStart is called when the TLS handshake is started. When
// connecting to an HTTPS site via an HTTP proxy, the handshake happens
// after the CONNECT request is processed by the proxy.
		TLSHandshakeStart: func() { tlsHandshake = time.Now() },
// TLSHandshakeDone is called after the TLS handshake with either the
// successful handshake's connection state, or a non-nil error on handshake
// failure.
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
		//log the time taken for TLS Handshake to complete in the log output using the logrus package. 
			logrus.Infof("TLS Handshake: %v", time.Since(tlsHandshake))
		},
// called when the HTTP client starts a new TCP connection to the server.
//ConnectStart function sets the connect variable to the current time using the time.Now() function
//log the network and addr parameters using the logrus.Debug() function.
//This allows for debugging and tracing of the TCP connection establishment process.
ConnectStart: func(network, addr string) {
			connect = time.Now()
			logrus.Debug(network, addr)
		},
		ConnectDone: func(network, addr string, err error) {
			logrus.Debug(network, addr, err)
			logrus.Infof("Connect time: %v", time.Since(connect))
		},

		GotFirstResponseByte: func() {
			logrus.Warnf("TAT : %v", time.Since(start))
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	start = time.Now()

	// NOTE: Below line is to ignore ssl certificate
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return res, err
	}
	return res, nil
}

