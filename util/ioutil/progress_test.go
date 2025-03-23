package ioutil_test

import (
	"bytes"
	"math/rand/v2"
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

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) {
	return f(p)
}

func TestProgressWriter_WriteString(t *testing.T) {
	var test, want bytes.Buffer
	var inputs = []string{
		"",
		" ",
		"	",
		"hello, world",
		"\a\b\\\\t\n\r\"'",
		"\x00\x01\x02\x03\x04",
	}
	pw1 := ioutil.NewProgressWriter(&test)
	pw2 := ioutil.NewProgressWriter(writerFunc(test.Write))
	for _, pw := range []*ioutil.ProgressWriter{pw1, pw2} {
		test.Reset()
		want.Reset()
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
}

func TestProgressWriter_Status(t *testing.T) {
	var SEED = uint64(time.Now().Unix())
	t.Logf("Running tests with PCG seed (0,%v)", SEED)
	rd := rand.New(rand.NewPCG(0, SEED))

	const SIZE = 32 * 1024 * 1024
	buf := make([]byte, SIZE)
	ioutil.ReadRand(rd, buf)

	pw := ioutil.NewProgressWriter(&bytes.Buffer{})
	if pw.Size() != 0 {
		t.Errorf("Size() = %v, want %v for new ProgressWriter", pw.Size(), 0)
	}

	go func() {
		defer pw.Close()
		pw.Write(buf)
	}()

	var last int64
	for n := range pw.Status() {
		if n < last {
			t.Errorf("Status() got %v, want larger than %v", n, last)
		}
		last = n
	}
	if last != pw.Size() || last != SIZE {
		t.Errorf("Status() got %v in the end, want %v", last, SIZE)
	}
}
