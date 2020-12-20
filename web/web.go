// REST frontend for the fooweb example

package web

import (
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"log"
	"sync"
	"time"
)

const (
	NATS_SUBJECT = "fooweb_req"
)

var (
	once sync.Once
	conn *nats.Conn
	counter int
)

func Connect() *nats.Conn {
	once.Do(func() {
		var err error
		conn, err = nats.Connect(nats.DefaultURL)
		if err != nil {
			log.Fatal(err)
		}
	})
	return conn
}

// Return the value of the request counter
func stats(c *gin.Context) {
	c.JSON(200, gin.H{
		"req_counter": counter,
	})
}

// Handle a request for the backend
func backend_req(c *gin.Context) {
	nc := Connect()
	filename := c.DefaultPostForm("filename", "/etc/hostname")

	resp, err := nc.Request(NATS_SUBJECT, []byte(filename), 2*time.Second)
	if err != nil {
		if nc.LastError() != nil {
			log.Fatalf("%v for request", nc.LastError())
		}
		log.Fatalf("%v for request", err)
	}
	counter++
	c.JSON(200, map[string]string{ "file_contents":
		string(resp.Data)})
}

// Setup the REST API with two endpoints
func Setup() *gin.Engine {
	r := gin.Default()
	r.GET("/stats", stats)
	r.POST("/req", backend_req)
	return r
}
