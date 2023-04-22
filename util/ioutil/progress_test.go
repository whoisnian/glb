package ioutil_test

import (
	"bytes"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/whoisnian/glb/util/ioutil"
)

func TestProgressWriter_Write(t *testing.T) {
	var test, want bytes.Buffer
	pw := ioutil.NewProgressWriter(&test)
	var inputs = [][]byte{
		nil,
		{},
		{'h', 'e', 'l', 'l', 'o'},
		{0, 1, 2, 3, 4},
	}
	for _, input := range inputs {
		n1, e1 := pw.Write(input)
		n2, e2 := want.Write(input)
		if n1 != n2 || e1 != e2 {
			t.Errorf("Write(%q) = %v %v, want %v %v", input, n1, e1, n2, e2)
		}
		if !bytes.Equal(test.Bytes(), want.Bytes()) {
			t.Errorf("bytes.Buffer.Bytes() = %v, want %v", test.Bytes(), want.Bytes())
		}
	}
}

func TestProgressWriter_WriteString(t *testing.T) {
	var test, want bytes.Buffer
	pw := ioutil.NewProgressWriter(&test)
	var inputs = []string{
		"",
		" ",
		"	",
		"hello, world",
		"\a\b\\\\t\n\r\"'",
		"\x00\x01\x02\x03\x04",
	}
	for _, input := range inputs {
		n1, e1 := pw.WriteString(input)
		n2, e2 := want.WriteString(input)
		if n1 != n2 || e1 != e2 {
			t.Errorf("WriteString(%q) = %v %v, want %v %v", input, n1, e1, n2, e2)
		}
		if !bytes.Equal(test.Bytes(), want.Bytes()) {
			t.Errorf("Bytes() = %v, want %v", test.Bytes(), want.Bytes())
		}
	}
}

func TestProgressWriter_Status(t *testing.T) {
	var SEED, SIZE int64 = time.Now().Unix(), 32 * 1024 * 1024
	t.Logf("Running tests with rand seed %v", SEED)

	rd := io.LimitReader(rand.New(rand.NewSource(SEED)), SIZE)
	pw := ioutil.NewProgressWriter(&bytes.Buffer{})

	if pw.Size() != 0 {
		t.Errorf("Size() = %v, want %v for new ProgressWriter", pw.Size(), 0)
	}

	go func() {
		defer pw.Close()
		io.Copy(pw, rd)
	}()

	last := 0
	for n := range pw.Status() {
		if n < last {
			t.Errorf("Status() got %v, want larger than %v", n, last)
		}
		last = n
	}
	if last != pw.Size() || last != int(SIZE) {
		t.Errorf("Status() got %v in the end, want %v", last, SIZE)
	}
}
