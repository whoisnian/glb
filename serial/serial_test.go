package serial_test

import (
	"fmt"
	"time"

	"github.com/whoisnian/glb/serial"
)

func ExampleOpen() {
	port, err := serial.Open("/dev/ttyUSB0", 115200, 8, serial.ParityNone, serial.StopBits1)
	if err != nil {
		panic(err)
	}
	port.SetInterval(time.Millisecond * 200)

	defer port.Close()
	stop := port.GoWaitAndSend()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, _ := port.Read(buf)
			for _, b := range buf[:n] {
				fmt.Printf("0x%02x ", b)
			}
			fmt.Printf("\n")
		}
	}()

	port.Push([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07})
	port.Push([]byte{0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f})

	time.Sleep(time.Second * 10)
	stop()
}
