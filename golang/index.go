package main

import (
	"flag"
	"os"
	"log"
	"net/http"
	"net/url"
	"io/ioutil"
	"encoding/json"
	"strings"
	"github.com/gorilla/websocket"
	JA3Trasport "github.com/CUCyber/ja3transport"
	// tls "github.com/refraction-networking/utls"
)

type myTLSRequest struct {
	RequestID string `json:"requestId"`
	Options struct {
		URL string `json:"url"`
		Method string `json:"method"`
		Headers map[string]string `json:"headers"`
		Body string `json:"body"`
		Ja3 string `json:"ja3"`
	} `json:"options"`
}

type response struct {
	Status int
	Body string
	Headers map[string]string
}

type myTLSResponse struct {
	RequestID string
	Response response 
}

func getWebsocketAddr() (string) {
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
			return
		}

		mytlsrequest := new(myTLSRequest)
		e := json.Unmarshal(message, &mytlsrequest)
		if e != nil {
			log.Print(err)
			return
		}

		client, err := JA3Trasport.NewWithString(string(mytlsrequest.Options.Ja3))
		if err != nil {
			log.Print(mytlsrequest.RequestID + "REQUESTIDONTHELEFT" + err.Error())
			return
		}

		req, err := http.NewRequest(strings.ToUpper(mytlsrequest.Options.Method), mytlsrequest.Options.URL, strings.NewReader(mytlsrequest.Options.Body))
		if err != nil {
			log.Print(err)
			return
		}

		for k, v := range mytlsrequest.Options.Headers {
			if k != "host" {
				req.Header.Set(k, v)
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Print(err)
			return
		}

		defer resp.Body.Close()
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Print(err)
			return
		}

		headers := make(map[string]string)

		for name, values := range resp.Header {
			for _, value := range values {		
				headers[name] = value
			}
		}

		Response := response{resp.StatusCode, string(bodyBytes), headers}

		reply := myTLSResponse{mytlsrequest.RequestID, Response}

		data, err := json.Marshal(reply)
		if err != nil {
			log.Print(err)
			return
		}
		
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Print(err)
			return
		}
	}
}