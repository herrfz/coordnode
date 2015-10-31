package app

import (
	"encoding/hex"
	"fmt"
	"github.com/herrfz/devreader"
	"github.com/tarm/goserial"
	"io"
	"os"
)

type SerialReader struct {
	serial io.ReadWriteCloser
}

func (s SerialReader) ReadDevice() ([]byte, error) { // TODO: check protocol
	buf := make([]byte, 128)
	msgLen, _ := s.serial.Read(buf)
	if msgLen > 0 {
		return buf[:msgLen], nil
	} else {
		return []byte{}, nil
	}
}

func DoForwardData(appDlCh, appUlCh chan []byte, device string) {
	siface := &serial.Config{Name: device, Baud: 9600} // TODO: check parameters
	serReader, err := serial.OpenPort(siface)
	if err != nil {
		fmt.Println("error opening serial interface:", err.Error())
		os.Exit(1)
	}
	defer serReader.Close()
	ser := SerialReader{serReader}
	serCh := devreader.MakeChannel(ser)

LOOP:
	for {
		select {
		case payload := <-serCh:
			appUlCh <- payload
			fmt.Println("forward sensor data:", hex.EncodeToString(payload))

		case _, more := <-appDlCh:
			if !more {
				close(appUlCh)
				break LOOP
			}
		}
	}
	fmt.Println("stopped forwarding sensor data")
}
