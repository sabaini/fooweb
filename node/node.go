package node

import (
	"github.com/nats-io/nats.go"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"
)

const (
	NATS_SUBJ  = "fooweb_req"
	NATS_QUEUE = "fooweb_queue"
)

// Worker function for the backend
func reply(msg *nats.Msg) {
	fn := string(msg.Data)
	content, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Printf("error reading %s: %s", fn, err)
	}
	if len(content) >= 1024*1024 {
		// Max. NATS message size
		content = content[:1024*1024]
	}
	_ = msg.Respond(content)
}

// Setup backend service
func SetupBackend() {
	opts := []nats.Option{nats.Name("fooweb backend")}
	opts = setupConnOptions(opts)

	// Connect to NATS, make conn encoded
	nc, err := nats.Connect(nats.DefaultURL, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	_, _ = nc.QueueSubscribe(NATS_SUBJ, NATS_QUEUE, reply) // errors handled below
	_ = nc.Flush()
	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on [%s]", NATS_SUBJ)

	// Setup the interrupt handler to drain so we don't miss
	// requests when scaling down.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c // Block
	log.Println()
	log.Printf("Draining...")
	_ = nc.Drain()
	log.Fatalf("Exiting")
}

func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		log.Printf("Disconnected due to: %s, will attempt reconnects for %.0fm", err, totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Fatalf("Exiting: %v", nc.LastError())
	}))
	return opts
}
