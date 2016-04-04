# REST to NATS proxy.

[![Build Status](https://travis-ci.org/sohlich/nats-proxy.svg?branch=master)](https://travis-ci.org/sohlich/nats-proxy)
[![GoReport](https://goreportcard.com/badge/github.com/sohlich/nats-proxy)](https://goreportcard.com/report/github.com/sohlich/nats-proxy)
[![GoDoc](https://godoc.org/github.com/sohlich/nats-proxy?status.svg)](https://godoc.org/github.com/sohlich/nats-proxy)




This library provides translation of HTTP requests to NATS messaging platform. 
Client side is partly inspired by https://github.com/gin-gonic/gin project to process the
requests.

#### The library is under development but the main API should remain stable.

### Example usage

#### service connected to NATS via nats-proxy client:
```
	clientConn, _ := nats.Connect(nats.DefaultURL)
	natsClient, _ := NewNatsClient(clientConn)
	natsClient.GET("/test/:event/:session", func(c *Context) {
		reqEvent = c.PathVariable("event")
		reqSession = c.PathVariable("session")

		respStruct := struct {
			User string
		}{
			"Radek",
		}

		c.JSON(200, respStruct)
	})
	defer clientConn.Close()
```
#### nats-proxy gateway
```
	proxyConn, _ := nats.Connect(nats.DefaultURL)
	proxy, _ := NewNatsProxy(proxyConn)
	http.Handle("/", proxy)
	defer proxyConn.Close()
```

#### HTTP client 
```
resp, err := http.Get("http://127.0.0.1:3000/test/12324/123")
```


### Advanced features

#### Proxy hook

Proxy hook is feature designed to change the
response before it's written to http response.
This feature could be used for example to 
enrich the header by special content or audit
outgoing data.

```
multiProxy.AddHook("/login.*", loginHook)

func loginHook(resp *natsproxy.Response) {
    // Do something 
    // with the response
    // e.g change outgoing header.
	resp.Header.Set(TokenHeader, token.RefToken)
}
```

#### Client middleware

The client middleware feature is inspired by gin framework.
Context could be processed before it is processed by specific handler.
The example below shows how to use middleware to log all incoming requests.

```
natsClient.Use(logger)

func logger(c *natsproxy.Context) {
	log.Infof("%s:%s from %s", c.Request.Method, c.Request.URL)
}
```






