package main

import (
	"flag"
	"fmt"
	work "github.com/herrfz/coordnode/worker"
	"github.com/herrfz/gowdc/utils"
	zmq "github.com/pebbe/zmq4"
	"os"
	"os/signal"
)

type Socket struct {
	socket *zmq.Socket
}

func (sock Socket) ReadDevice() ([]byte, error) {
	buf, err := sock.socket.Recv(0)
	if err != nil {
		if err.Error() == "Operation cannot be accomplished in current state" {
			// give some time for REP socket to send before another recv
			return nil, fmt.Errorf("DONTPANIC")
		}
	}
	return []byte(buf), err
}

func main() {
	serial := flag.Bool("serial", false, "use serial port to talk to sensor node")
	device := flag.String("device", "", "serial device to use")
	flag.Parse()

	if *serial && *device == "" {
		fmt.Println("no serial device provided")
		os.Exit(1)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	c_sock, _ := zmq.NewSocket(zmq.REP)
	defer c_sock.Close()
	c_sock.Bind("tcp://*:5555")

	d_dl_sock, _ := zmq.NewSocket(zmq.PULL)
	defer d_dl_sock.Close()
	d_dl_sock.Connect("tcp://localhost:5556")

	d_ul_sock, _ := zmq.NewSocket(zmq.PUSH)
	defer d_ul_sock.Close()
	d_ul_sock.Bind("tcp://*:5557")

	dl_chan := make(chan []byte)
	ul_chan := make(chan []byte)

	if *serial {
		go work.DoSerial(dl_chan, ul_chan, *device)
	} else {
		go work.DoEmulCoordNode(dl_chan, ul_chan)
	}

	data_ch := utils.MakeChannel(Socket{d_dl_sock})
	cmd_ch := utils.MakeChannel(Socket{c_sock})

LOOP:
	for {
		select {
		case buf := <-cmd_ch:
			respmsg := work.ProcessMessage([]byte(buf))
			if respmsg != nil {
				c_sock.Send(string(respmsg), 0)
				fmt.Println("sent answer to TCP command")
			}

		case buf := <-data_ch:
			respmsg := work.ProcessMessage([]byte(buf))
			if respmsg != nil {
				d_ul_sock.Send(string(respmsg), 0)
				fmt.Println("sent answer to UDP mcast message")
			}
			dl_chan <- []byte(buf)

		case buf := <-ul_chan:
			d_ul_sock.Send(string(buf), 0)
			fmt.Println("sent node uplink message")

		case <-c:
			close(dl_chan)
			<-ul_chan
			break LOOP
		}
	}
	fmt.Println("Program stopped")
}
