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

type myTLSRequest struct {
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

type response struct {
	Status  int
	Body    string
	Headers map[string]string
}

type myTLSResponse struct {
	RequestID string
	Response  response
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

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Print(err)
			continue
		}

		mytlsrequest := new(myTLSRequest)
		e := json.Unmarshal(message, &mytlsrequest)
		if e != nil {
			log.Print(err)
			continue
		}

		config := &tls.Config{
			InsecureSkipVerify: true,
		}

		var transport http.RoundTripper

		rawProxy := mytlsrequest.Options.Proxy
		if rawProxy != "" {
			proxyURL, _ := url.Parse(rawProxy)
			proxy, err := FromURL(proxyURL, Direct)
			if err != nil {
				log.Print(mytlsrequest.RequestID + "Request_Id_On_The_Left" + err.Error())
				continue
			}

			tr, err := NewTransportWithDialer(string(mytlsrequest.Options.Ja3), config, proxy)
			if err != nil {
				log.Print(mytlsrequest.RequestID + "Request_Id_On_The_Left" + err.Error())
				continue
			}
			transport = tr

		} else {
			tr, err := NewTransportWithConfig(string(mytlsrequest.Options.Ja3), config)
			if err != nil {
				log.Print(mytlsrequest.RequestID + "Request_Id_On_The_Left" + err.Error())
				continue
			}
			transport = tr
		}

		client := &http.Client{Transport: transport}

		req, err := http.NewRequest(strings.ToUpper(mytlsrequest.Options.Method), mytlsrequest.Options.URL, strings.NewReader(mytlsrequest.Options.Body))
		if err != nil {
			log.Print(mytlsrequest.RequestID + "Request_Id_On_The_Left" + err.Error())
			continue
		}

		for k, v := range mytlsrequest.Options.Headers {
			if k != "host" {
				req.Header.Set(k, v)
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Print(mytlsrequest.RequestID + "Request_Id_On_The_Left" + err.Error())
			continue
		}

		defer resp.Body.Close()
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Print(mytlsrequest.RequestID + "Request_Id_On_The_Left" + err.Error())
			continue
		}

		headers := make(map[string]string)

		for name, values := range resp.Header {
			if name == "Set-Cookie" {
				headers[name] = strings.Join(values, "/,/")
			} else {
				for _, value := range values {
					headers[name] = value
				}
			}
		}

		Response := response{resp.StatusCode, string(bodyBytes), headers}

		reply := myTLSResponse{mytlsrequest.RequestID, Response}

		data, err := json.Marshal(reply)
		if err != nil {
			log.Print(mytlsrequest.RequestID + "Request_Id_On_The_Left" + err.Error())
			continue
		}

		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Print(mytlsrequest.RequestID + "Request_Id_On_The_Left" + err.Error())
			continue
		}
	}
}
