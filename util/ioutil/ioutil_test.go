package ioutil_test

import (
	"bytes"
	"errors"
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
	defer f.Close()
	defer os.Remove(f.Name())

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
