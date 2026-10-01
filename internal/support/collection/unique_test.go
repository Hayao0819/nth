package collection

import (
	"reflect"
	"testing"
)

func TestAppendUniqueBy(t *testing.T) {
	t.Parallel()

	type item struct {
		id    string
		value string
	}
	got := AppendUniqueBy(
		[]item{{id: "1", value: "first"}, {id: "2", value: "second"}},
		[]item{{id: "2", value: "replacement"}, {id: "3", value: "third"}},
		func(value item) string { return value.id },
	)
	want := []item{{id: "1", value: "first"}, {id: "2", value: "second"}, {id: "3", value: "third"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("AppendUniqueBy() = %#v, want %#v", got, want)
	}
}
