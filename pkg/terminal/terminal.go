package terminal

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/whoisnian/glb/pkg/ansi"
	"golang.org/x/sys/unix"
)

type Terminal struct {
	fd   int
	lock *sync.Mutex

	MaxLine int
}

func (terminal *Terminal) Lock() {
	terminal.lock.Lock()
}

func (terminal *Terminal) Unlock() {
	terminal.lock.Unlock()
}

func (terminal *Terminal) WriteString(s string) (n int, err error) {
	terminal.lock.Lock()
	defer terminal.lock.Unlock()

	return os.Stdout.WriteString(s)
}

func (terminal *Terminal) GetSize() (width, height int, err error) {
	ws, err := unix.IoctlGetWinsize(terminal.fd, unix.TIOCGWINSZ)
	if err != nil {
		return -1, -1, err
	}
	return int(ws.Col), int(ws.Row), nil
}

func (terminal *Terminal) Clear() {
	terminal.WriteString(ansi.ClearScreen)
}

func (terminal *Terminal) ScrollClear() (err error) {
	if _, winHeight, err := terminal.GetSize(); err == nil {
		if cursorRow, _, err := terminal.GetCursorPos(); err == nil {
			terminal.WriteString(fmt.Sprintf(ansi.SetCursorAddress, winHeight, 1))
			terminal.ScrollDown(cursorRow)
			terminal.WriteString(ansi.ClearScreen)
		}
	}
	return
}

func (terminal *Terminal) ClearLineToRight() {
	terminal.WriteString(ansi.ClearLineToRight)
}

func (terminal *Terminal) ScrollUp(n int) {
	terminal.WriteString(strings.Repeat(ansi.ScrollUp, n))
}

func (terminal *Terminal) ScrollDown(n int) {
	terminal.WriteString(strings.Repeat(ansi.ScrollDown, n))
}

func (terminal *Terminal) GetCursorPos() (row, col int, err error) {
	termios, err := unix.IoctlGetTermios(terminal.fd, unix.TCGETS)
	if err != nil {
		return -1, -1, err
	}
	terminal.WriteString(ansi.GetCursorAddress)

	newState := *termios
	newState.Lflag &^= unix.ECHO | unix.ICANON
	if err = unix.IoctlSetTermios(terminal.fd, unix.TCSETS, &newState); err != nil {
		return -1, -1, err
	}
	defer unix.IoctlSetTermios(terminal.fd, unix.TCSETS, termios)

	var buf [1]byte
	var res []byte
	for {
		n, err := unix.Read(terminal.fd, buf[:])
		if n > 0 {
			if buf[0] == 'R' {
				break
			}
			res = append(res, buf[0])
		}
		if err != nil {
			break
		}
	}

	pos := bytes.Split(res[2:], []byte{';'})
	if len(pos) < 2 {
		return -1, -1, err
	}
	row, _ = strconv.Atoi(string(pos[0]))
	col, _ = strconv.Atoi(string(pos[1]))
	return row - 1, col - 1, nil
}

func (terminal *Terminal) SetCursorPos(row, col int) {
	if row > terminal.MaxLine {
		terminal.MaxLine = row
	}
	terminal.WriteString(fmt.Sprintf(ansi.SetCursorAddress, row+1, col+1))
}

func (terminal *Terminal) MoveCursorToMaxLineNext() {
	terminal.WriteString(fmt.Sprintf(ansi.SetCursorAddress, terminal.MaxLine+2, 1))
}

func (terminal *Terminal) ShowCursor() {
	terminal.WriteString(ansi.CursorVisible)
}

func (terminal *Terminal) HideCursor() {
	terminal.WriteString(ansi.CursorInvisible)
}

func (terminal *Terminal) EnableLineWrap() {
	terminal.WriteString(ansi.EnableLineWrap)
}

func (terminal *Terminal) DisableLineWrap() {
	terminal.WriteString(ansi.DisableLineWrap)
}

func IsTerminal() bool {
	_, err := unix.IoctlGetTermios(int(os.Stdin.Fd()), unix.TCGETS)
	return err == nil
}

func New() *Terminal {
	return &Terminal{
		int(os.Stdin.Fd()),
		new(sync.Mutex),
		0,
	}
}
