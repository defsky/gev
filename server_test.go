package gev

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/Allenxuxu/gev/connection"
	"github.com/Allenxuxu/ringbuffer"
	"github.com/Allenxuxu/toolkit/sync"
)

type example struct{}

func (s *example) OnConnect(c *connection.Connection) {
	//log.Println(" OnConnect ： ", c.PeerAddr())
}
func (s *example) OnMessage(c *connection.Connection, buffer *ringbuffer.RingBuffer) (out []byte) {
	//log.Println("OnMessage")
	outbuf := buffer.Bytes()
	buffer.RetrieveAll()

	_ = c.Send(outbuf)
	return
}

func (s *example) OnClose(c *connection.Connection) {
	//log.Println("OnClose")
}

func TestServer_Start(t *testing.T) {
	handler := new(example)

	s, err := NewServer(handler,
		Network("tcp"),
		Address(":1833"),
		NumLoops(8),
		ReusePort(true))
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(time.Second)
		sw := sync.WaitGroupWrapper{}
		for i := 0; i < 100; i++ {
			sw.AddAndRun(func() {
				startClient(s.opts.Network, s.opts.Address)
			})
		}

		sw.Wait()
		log.Println("stop server")
		s.Stop()
	}()

	s.Start()
}

func startClient(network, addr string) {
	rand.Seed(time.Now().UnixNano())
	c, err := net.Dial(network, addr)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	rd := bufio.NewReader(c)
	duration := time.Duration((rand.Float64()*2+1)*float64(time.Second)) / 8
	start := time.Now()
	for time.Since(start) < duration {
		sz := rand.Int() % (1024 * 1024)
		data := make([]byte, sz)
		if _, err := rand.Read(data); err != nil {
			panic(err)
		}
		if _, err := c.Write(data); err != nil {
			panic(err)
		}
		data2 := make([]byte, len(data))
		if _, err := io.ReadFull(rd, data2); err != nil {
			panic(err)
		}
		if string(data) != string(data2) {
			panic("mismatch")
		}
	}
}

func ExampleServer_RunAfter() {
	handler := new(example)

	s, err := NewServer(handler,
		Network("tcp"),
		Address(":1833"),
		NumLoops(8),
		ReusePort(true))
	if err != nil {
		panic(err)
	}

	go s.Start()
	defer s.Stop()

	s.RunAfter(time.Second, func() {
		fmt.Println("RunAfter")
	})

	time.Sleep(2500 * time.Millisecond)

	// Output:
	// RunAfter
}

func ExampleServer_RunEvery() {
	handler := new(example)

	s, err := NewServer(handler,
		Network("tcp"),
		Address(":1833"),
		NumLoops(8),
		ReusePort(true))
	if err != nil {
		panic(err)
	}

	go s.Start()
	defer s.Stop()

	t := s.RunEvery(time.Second, func() {
		fmt.Println("EveryFunc")
	})

	time.Sleep(4500 * time.Millisecond)
	t.Stop()
	time.Sleep(4500 * time.Millisecond)

	// Output:
	// EveryFunc
	// EveryFunc
	// EveryFunc
	// EveryFunc
}
