package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"strings"
	"log"
	"net/http"
	"net/url"
	"os"
	tls "github.com/refraction-networking/utls"
	// JA3 "github.com/CUCyber/ja3transport"
)

type MyTlsRequest struct {
	RequestID string `json:"requestId"`
	Options   struct {
		URL     string            `json:"url"`
		Method  string            `json:"method"`
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
		Ja3     string            `json:"ja3"`
		Proxy   string            `json:"proxy"`
	} `json:"options"`
}

type RequestResponse struct {
	Status  int
	Body    string
	Headers map[string]string
}

type MyTlsResponse struct {
	RequestID string
	Response  RequestResponse
}

func getWebsocketAddr() string {
	port, exists := os.LookupEnv("WS_PORT")

	var addr *string

	if exists {
		addr = flag.String("addr", "localhost:"+port, "http service address")
	} else {
		addr = flag.String("addr", "localhost:9119", "http service address")
	}

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}

	return u.String()
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	websocketAddress := getWebsocketAddr()

	c, _, err := websocket.DefaultDialer.Dial(websocketAddress, nil)
	if err != nil {
		log.Print(err)
		return
	}

	// TODO: move all definitions out of infinite loop, memory leak
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("error reading ws message:", err)
			continue
		}

		tlsRequest := new(MyTlsRequest)
		e := json.Unmarshal(message, &tlsRequest)
		if e != nil {
			log.Println("error unmarshalling request json:", err)
			continue
		}

		config := &tls.Config{
			InsecureSkipVerify: true,
		}

		var transport http.RoundTripper

		rawProxy := tlsRequest.Options.Proxy
		if rawProxy != "" {
			proxyURL, _ := url.Parse(rawProxy)
			proxy, err := FromURL(proxyURL, Direct)
			if err != nil {
				log.Printf("[%s] error parsing proxy url: %s\n", tlsRequest.RequestID, err)
				continue
			}

			tr, err := NewTransportWithDialer(tlsRequest.Options.Ja3, config, proxy)
			if err != nil {
				log.Printf("[%s] error creating transport: %s\n", tlsRequest.RequestID, err)
				continue
			}
			transport = tr

		} else {
			tr, err := NewTransportWithConfig(tlsRequest.Options.Ja3, config)
			if err != nil {
				log.Printf("[%s] error creating transport: %s\n", tlsRequest.RequestID, err)
				continue
			}
			transport = tr
		}

		client := &http.Client{Transport: transport}

		req, err := http.NewRequest(strings.ToUpper(tlsRequest.Options.Method), tlsRequest.Options.URL, strings.NewReader(tlsRequest.Options.Body))
		if err != nil {
			log.Printf("[%s] error creating request: %s\n", tlsRequest.RequestID, err)
			continue
		}

		for k, v := range tlsRequest.Options.Headers {
			// TODO: reconsider this check for 2 reasons,
			// 1st we should trust that the correct host header is provided if it is provided at all
			// and 2nd it doesn't even work if they name the header with any capital letters
			if k != "host" {
				req.Header.Set(k, v)
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[%s] error performing request: %s\n", tlsRequest.RequestID, err)
			continue
		}

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("[%s] error reading response body: %s\n", tlsRequest.RequestID, err)
			continue
		}

		headers := make(map[string]string)

		// TODO: better header handling. there are blatant issues with this method of handling headers.
		for name, values := range resp.Header {
			if name == "Set-Cookie" {
				headers[name] = strings.Join(values, "/,/")
			} else {
				headers[name] = values[len(values)-1]
			}
		}

		Response := RequestResponse{resp.StatusCode, string(bodyBytes), headers}

		reply := MyTlsResponse{tlsRequest.RequestID, Response}

		data, err := json.Marshal(reply)
		if err != nil {
			log.Printf("[%s] error marshalling reply json: %s\n", tlsRequest.RequestID, err)
			continue
		}

		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("[%s] error writing message to ws: %s\n", tlsRequest.RequestID, err)
			continue
		}
	}
}
