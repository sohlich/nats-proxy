# REST to NATS proxy.

[![Build Status](https://travis-ci.org/sohlich/nats-proxy.svg?branch=master)](https://travis-ci.org/sohlich/nats-proxy)


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
	proxyHandler, _ := NewNatsProxy(proxyConn)
	http.Handle("/", proxyHandler)
	defer proxyConn.Close()
```

#### HTTP client 
```
resp, err := http.Get("http://127.0.0.1:3000/test/12324/123")
```