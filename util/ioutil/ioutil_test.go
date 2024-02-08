package ioutil_test

import (
	"bytes"
	"errors"
	"math/rand/v2"
	"os"
	"testing"

	"github.com/whoisnian/glb/util/ioutil"
)

func TestSeekAndReadAll(t *testing.T) {
	if _, err := ioutil.SeekAndReadAll(nil); !errors.Is(err, os.ErrInvalid) {
		t.Fatalf("SeekAndReadAll(nil) should return os.ErrInvalid")
	}

	f, err := os.CreateTemp(os.TempDir(), "TestSeekAndReadAll_")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	checkEqualWant := func(want []byte) {
		if data, err := ioutil.SeekAndReadAll(f); err != nil {
			t.Fatalf("SeekAndReadAll: %v", err)
		} else if !bytes.Equal(data, want) {
			t.Fatalf("SeekAndReadAll return %q, want %q", data, want)
		}
	}

	want := []byte{}
	checkEqualWant(want)

	testData := []byte("TestSeekAndReadAll")
	f.Write(testData)
	want = testData
	checkEqualWant(want)
	checkEqualWant(want)

	f.Write(testData)
	want = append(testData, testData...)
	checkEqualWant(want)
	checkEqualWant(want)
}

func TestSeekAndReadAllProc(t *testing.T) {
	// from os.TestReadFileProc()
	name := "/proc/sys/fs/pipe-max-size"
	if _, err := os.Stat(name); err != nil {
		t.Skip(err)
	}

	f, err := os.Open(name)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	checkAndReturn := func() []byte {
		data, err := ioutil.SeekAndReadAll(f)
		if err != nil {
			t.Fatalf("SeekAndReadAll: %v", err)
		} else if len(data) == 0 || data[len(data)-1] != '\n' {
			t.Fatalf("read %s: not newline-terminated: %q", name, data)
		}
		return data
	}

	data1 := checkAndReturn()
	data2 := checkAndReturn()
	if !bytes.Equal(data1, data2) {
		t.Fatalf("%s content should not change: %q %q", name, data1, data2)
	}
}

func TestReadRand(t *testing.T) {
	// from rand.TestPCG()
	rd := rand.New(rand.NewPCG(1, 2))
	var tests = []struct {
		size int
		want []byte
	}{
		{0, []byte{}},                       // nop
		{1, []byte{0x10}},                   // 0xc4f5a58656eef510
		{2, []byte{0x6c, 0xec}},             // 0x9dcec3ad077dec6c
		{3, []byte{0x88, 0x80, 0x2f}},       // 0xc8d04605312f8088
		{4, []byte{0x9a, 0xc1, 0x3a, 0xb6}}, // 0xcbedc0dcb63ac19a
		{8, []byte{0x50, 0x79, 0xe9, 0xca, 0x98, 0x87, 0xf9, 0x3b}}, // 0x3bf98798cae97950
		{10, []byte{
			0xbc, 0x5a, 0x48, 0x8d, 0x7f, 0x6d, 0x8c, 0x0a, // 0xa8c6d7f8d485abc
			0x79, 0xd2, // 0x7ffa3780429cd279
		}},
		{16, []byte{
			0x8e, 0x2f, 0x1c, 0x6b, 0x62, 0xd2, 0x0a, 0x73, // 0x730ad2626b1c2f8e
			0x99, 0xad, 0xa0, 0xf4, 0x30, 0x23, 0xff, 0x21, // 0x21ff2330f4a0ad99
		}},
		{20, []byte{
			0xb0, 0x94, 0x70, 0x94, 0xa1, 0x01, 0x09, 0x2f, // 0x2f0901a1947094b0
			0xef, 0x6c, 0xe3, 0xfb, 0x3c, 0x5a, 0x73, 0xa9, // 0xa9735a3cfbe36cef
			0x4a, 0xc8, 0x12, 0x1a, // 0x71ddb0a01a12c84a
		}},
		{32, []byte{
			0xbb, 0x53, 0x84, 0xa7, 0x77, 0x3e, 0xe5, 0xf0, // 0xf0e53e77a78453bb
			0x9d, 0x1e, 0xbe, 0x63, 0x96, 0x3e, 0x17, 0x1f, // 0x1f173e9663be1e9d
			0x5e, 0x11, 0xc4, 0x3a, 0xda, 0x51, 0x76, 0x65, // 0x657651da3ac4115e
			0x7b, 0x15, 0x5a, 0xb6, 0x76, 0x73, 0x98, 0xc8, // 0xc8987376b65a157b
		}},
	}
	buf := make([]byte, 32)
	for _, test := range tests {
		ioutil.ReadRand(rd, buf[:test.size])
		if !bytes.Equal(buf[:test.size], test.want) {
			t.Errorf("ReadRand(%d) = %v, want %v", test.size, buf[:test.size], test.want)
		}
	}
}
