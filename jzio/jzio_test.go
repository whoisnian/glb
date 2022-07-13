package jzio_test

import (
	"bytes"
	"encoding/base64"
	"reflect"
	"testing"

	"github.com/whoisnian/glb/jzio"
)

type jzioTestT1 struct {
	A int
	B float64
	C string
}

type jzioTestT2 struct {
	ABC jzioTestT1
	D   []int
	E   []byte
	F   [][]byte
}

type jzioTestT3 struct {
	DEF jzioTestT2
	G   map[string]int
	H   map[int]bool
	I   map[string][]byte
}

type jzioTestInput struct {
	v any // value
	p any // pointer
}

var jzioTests = []struct {
	inputs []jzioTestInput
	resB64 string
}{{
	[]jzioTestInput{
		{-1, new(int)},
		{-0.5, new(float64)},
		{0, new(int)},
		{0.5, new(float64)},
		{1, new(int)},
		{65536, new(int)},
	},
	"0jXkAgAAAP//0jXQM+UCAAAA//8y4AIAAAD//wKRAAAAAP//MuQCAAAA//8yMzU1NuMCAAAA//8BAAD//w==",
}, {
	[]jzioTestInput{
		{nil, new(any)},
		{true, new(bool)},
		{"", new(string)},
		{"hello", new(string)},
		{"测试", new(string)},
	},
	"yivNyeECAAAA//8qKSpN5QIAAAD//1JS4gIAAAD//1LKSM3JyVfiAgAAAP//Unq2tfvF+qlKXAAAAAD//wEAAP//",
}, {
	[]jzioTestInput{
		{[]int{1, 2, 3}, new([]int)},
		{[]string{"a", "b", "c"}, new([]string)},
		{[]byte{0x01, 0x02, 0x03}, new([]byte)},
		{[]bool{true, false}, new([]bool)},
	},
	"ijbUMdIxjuUCAAAA//+KVkpU0lFKUtJRSlaK5QIAAAD//1JyDPR0UeICAAAA//+KLikqTdVJS8wpTo3lAgAAAP//AQAA//8=",
}, {
	[]jzioTestInput{
		{map[string]int{"a": 1, "b": 2}, new(map[string]int)},
		{map[string]string{"a": "s1", "b": "s2"}, new(map[string]string)},
		{map[int]bool{1: true, 2: false}, new(map[int]bool)},
	},
	"qlZKVLIy1FFKUrIyquUCAAAA//8C85SKDZXAAkrFRkq1XAAAAAD//6pWMlSyKikqTdVRMlKySkvMKU6t5QIAAAD//wEAAP//",
}, {
	[]jzioTestInput{
		{map[int][]int{1: {0}, 2: {2, 4}}, new(map[int][]int)},
		{map[string][]string{"a": {"A"}, "b": {"B", "C"}}, new(map[string][]string)},
		{map[int][]byte{1: {0x01}, 2: {0x02, 0x03}}, new(map[int][]byte)},
	},
	"qlYyVLKKNojVUTJSsoo20jGJreUCAAAA//+qVkpUsopWclSK1VFKArGclHSUnJVia7kAAAAA//8Cq1ZyDLS1VQJrUHJM97VVquUCAAAA//8BAAD//w==",
}, {
	[]jzioTestInput{
		{jzioTestT1{1, 0.5, "ab"}, new(jzioTestT1)},
		{jzioTestT2{
			jzioTestT1{1, 0.5, "ab"},
			[]int{1, 2},
			[]byte{0x01, 0x02},
			[][]byte{{0x01}, {0x02, 0x03}},
		}, new(jzioTestT2)},
		{jzioTestT3{
			jzioTestT2{
				jzioTestT1{1, 0.5, "ab"},
				[]int{1, 2},
				[]byte{0x01, 0x02},
				[][]byte{{0x01}, {0x02, 0x03}},
			},
			map[string]int{"a": 1, "b": 2},
			map[int]bool{1: true, 2: false},
			map[string][]byte{"a": {0x01}, "b": {0x02, 0x03}},
		}, new(jzioTestT3)},
	},
	"qlZyVLIy1FFyUrIy0DPVUXJWslJKTFKq5QIAAAD//6pWcnRyVrLCJa+j5KJkFW2oYxSro+SqZKXkGOhpq6Sj5KZkFa3kGGgLYjum+9oqxdZyAQAAAP//qlZycXUDm0QdA3WU3EHGJIKNSVKyMqrVUfIAiRgqWZUUlabqKBkpWaUl5hSn1uooeUKVwsxIAjFB5tTWcgEAAAD//wEAAP//",
}}

func TestReader(t *testing.T) {
	for i, test := range jzioTests {
		resBytes, err := base64.StdEncoding.DecodeString(test.resB64)
		if err != nil {
			t.Errorf("#%d. Base64 decode error: %v", i, err)
		}
		r := jzio.NewReader(bytes.NewBuffer(resBytes))
		defer r.Close()
		for j, input := range test.inputs {
			// reflect usage from:
			// https://github.com/golang/go/blob/4aa1efed4853ea067d665a952eee77c52faac774/src/encoding/json/decode_test.go#L1091
			typ := reflect.TypeOf(input.p)
			if typ.Kind() != reflect.Pointer {
				t.Errorf("#%d.%d. inputs[%d].p is not a pointer", i, j, j)
				break
			}
			v := reflect.New(typ.Elem())
			if err := r.UnMarshal(v.Interface()); err != nil {
				t.Errorf("#%d.%d Reader.UnMarshal() error: %v", i, j, err)
			} else if !reflect.DeepEqual(v.Elem().Interface(), input.v) {
				t.Errorf("#%d.%d Reader.UnMarshal() = %T, want %T", i, j, v.Elem().Interface(), input.v)
			}
		}
	}
}

func TestWriter(t *testing.T) {
	for i, test := range jzioTests {
		buf := new(bytes.Buffer)
		w, err := jzio.NewWriter(buf)
		if err != nil {
			t.Errorf("#%d. NewWriter() error: %v", i, err)
		}
		for j, input := range test.inputs {
			if err := w.Marshal(input.v); err != nil {
				t.Errorf("#%d.%d Writer.Marshal() error: %v", i, j, err)
			}
		}
		if err := w.Close(); err != nil {
			t.Errorf("#%d. Writer.Close() error: %v", i, err)
		}
		got := base64.StdEncoding.EncodeToString(buf.Bytes())
		if test.resB64 != got {
			t.Errorf("#%d. Writer.Marshal() = %v, want %v", i, got, test.resB64)
		}
	}
}

func TestReadWriter(t *testing.T) {
	for i, test := range jzioTests {
		resBytes, err := base64.StdEncoding.DecodeString(test.resB64)
		if err != nil {
			t.Errorf("#%d. Base64 decode error: %v", i, err)
		}
		buf := new(bytes.Buffer)
		rw, err := jzio.NewReadWriter(bytes.NewBuffer(resBytes), buf)
		if err != nil {
			t.Errorf("#%d. NewReadWriter() error: %v", i, err)
		}
		for j, input := range test.inputs {
			typ := reflect.TypeOf(input.p)
			if typ.Kind() != reflect.Pointer {
				t.Errorf("#%d.%d. inputs[%d].p is not a pointer", i, j, j)
				break
			}
			v := reflect.New(typ.Elem())
			if err := rw.UnMarshal(v.Interface()); err != nil {
				t.Errorf("#%d.%d ReadWriter.UnMarshal() error: %v", i, j, err)
			} else if !reflect.DeepEqual(v.Elem().Interface(), input.v) {
				t.Errorf("#%d.%d ReadWriter.UnMarshal() = %T, want %T", i, j, v.Elem().Interface(), input.v)
			} else if err := rw.Marshal(v.Elem().Interface()); err != nil {
				t.Errorf("#%d.%d ReadWriter.Marshal() error: %v", i, j, err)
			}
		}
		if err := rw.Close(); err != nil {
			t.Errorf("#%d. ReadWriter.Close() error: %v", i, err)
		}
		got := base64.StdEncoding.EncodeToString(buf.Bytes())
		if test.resB64 != got {
			t.Errorf("#%d. ReadWriter.Marshal() = %v, want %v", i, got, test.resB64)
		}
	}
}
