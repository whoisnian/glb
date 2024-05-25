package config_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/whoisnian/glb/config"
)

type PrefilledSimple struct {
	IntEmpty    int
	IntKeep     int
	IntOverride int

	StringEmpty    string
	StringKeep     string
	StringOverride string

	DurationEmpty    time.Duration
	DurationKeep     time.Duration
	DurationOverride time.Duration

	BytesEmpty    []byte
	BytesKeep     []byte
	BytesOverride []byte
}

func TestPrefilled_Simple(t *testing.T) {
	input := PrefilledSimple{
		IntKeep:          10,
		IntOverride:      10,
		StringKeep:       "hello",
		StringOverride:   "hello",
		DurationKeep:     time.Second,
		DurationOverride: time.Second,
		BytesKeep:        []byte("_nian_"),
		BytesOverride:    []byte("_nian_"),
	}
	data := []byte(`{"IntOverride":20,"StringOverride":"world","DurationOverride":60000000000,"BytesOverride":"X3h4eHhf"}`)
	want := PrefilledSimple{
		IntKeep:          10,
		IntOverride:      20,
		StringKeep:       "hello",
		StringOverride:   "world",
		DurationKeep:     time.Second,
		DurationOverride: time.Minute,
		BytesKeep:        []byte("_nian_"),
		BytesOverride:    []byte("_xxxx_"),
	}

	if err := config.JsonUnmarshal(data, &input); err != nil {
		t.Fatalf("config.JsonUnmarshal() error: %v", err)
	}
	if !reflect.DeepEqual(input, want) {
		t.Fatalf("config.JsonUnmarshal() result:\n  get  %+v\n  want %+v", input, want)
	}
}

type IntBox struct {
	IntEmpty    int
	IntKeep     int
	IntOverride int
}

type IntBox2 struct {
	BoxEmpty    IntBox
	BoxKeep     IntBox
	BoxOverride IntBox
}

type PrefilledNested struct {
	BoxEmpty    IntBox
	BoxKeep     IntBox
	BoxOverride IntBox

	Box2Empty    IntBox2
	Box2Keep     IntBox2
	Box2Override IntBox2
}

func TestPrefilled_Nested(t *testing.T) {
	input := PrefilledNested{
		BoxKeep:     IntBox{IntKeep: 10, IntOverride: 10},
		BoxOverride: IntBox{IntKeep: 10, IntOverride: 10},
		Box2Keep: IntBox2{
			BoxKeep:     IntBox{IntKeep: 10, IntOverride: 10},
			BoxOverride: IntBox{IntKeep: 10, IntOverride: 10},
		},
		Box2Override: IntBox2{
			BoxKeep:     IntBox{IntKeep: 10, IntOverride: 10},
			BoxOverride: IntBox{IntKeep: 10, IntOverride: 10},
		},
	}
	data := []byte(`{"BoxOverride":{"IntOverride":20},"Box2Override":{"BoxOverride":{"IntOverride":20}}}`)
	want := PrefilledNested{
		BoxKeep:     IntBox{IntKeep: 10, IntOverride: 10},
		BoxOverride: IntBox{IntKeep: 10, IntOverride: 20},
		Box2Keep: IntBox2{
			BoxKeep:     IntBox{IntKeep: 10, IntOverride: 10},
			BoxOverride: IntBox{IntKeep: 10, IntOverride: 10},
		},
		Box2Override: IntBox2{
			BoxKeep:     IntBox{IntKeep: 10, IntOverride: 10},
			BoxOverride: IntBox{IntKeep: 10, IntOverride: 20},
		},
	}

	if err := config.JsonUnmarshal(data, &input); err != nil {
		t.Fatalf("config.JsonUnmarshal() error: %v", err)
	}
	if !reflect.DeepEqual(input, want) {
		t.Fatalf("config.JsonUnmarshal() result:\n  get  %+v\n  want %+v", input, want)
	}
}
