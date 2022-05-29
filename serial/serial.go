//go:build linux

// Package serial implements a simple serial port descriptor for linux.
// from: https://github.com/tarm/serial
package serial

import (
	"os"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

type Port struct {
	*os.File
	wg       *sync.WaitGroup
	buf      chan []byte
	interval time.Duration
}

// Example
//   device: /dev/ttyUSB0
//   baudrate: 9600 (9600|19200|38400|57600|115200)
//   databits: 8 (5|6|7|8)
//   parity: ParityNone (ParityNone|ParityOdd|ParityEven|ParityMark|ParitySpace)
//   stopbits: StopBits1 (StopBits1|StopBits2)
func Open(device string, baudrate int, databits int, parity uint32, stopbits uint32) (p *Port, err error) {
	baudrateValid, ok := baudrateMap[baudrate]
	if !ok {
		return nil, ErrInvalidBaudRate
	}
	databitsValid, ok := databitsMap[databits]
	if !ok {
		return nil, ErrInvalidDataBits
	}

	f, err := os.OpenFile(device, unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0666)
	if err != nil {
		return nil, err
	}

	state := unix.Termios{
		Iflag:  unix.IGNPAR,
		Cflag:  unix.CREAD | unix.CLOCAL | baudrateValid | databitsValid | parity | stopbits,
		Ispeed: baudrateValid,
		Ospeed: baudrateValid,
	}
	state.Cc[unix.VMIN] = 1
	state.Cc[unix.VTIME] = 0
	if err = unix.IoctlSetTermios(int(f.Fd()), unix.TCSETS, &state); err != nil {
		return nil, err
	}

	if err = unix.SetNonblock(int(f.Fd()), false); err != nil {
		return nil, err
	}

	return &Port{
		File:     f,
		wg:       new(sync.WaitGroup),
		buf:      make(chan []byte, 256),
		interval: 0,
	}, nil
}

func (p *Port) SetInterval(interval time.Duration) {
	p.interval = interval
}

func (p *Port) Push(data []byte) {
	p.buf <- data
}

func (p *Port) GoWaitAndSend() (stop func()) {
	p.wg.Add(1)
	go func() {
		for data := range p.buf {
			p.Write(data)
			time.Sleep(p.interval)
		}
	}()

	return func() {
		close(p.buf)
		p.wg.Done()
	}
}
