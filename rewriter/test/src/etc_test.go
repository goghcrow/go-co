package src

import (
	"reflect"
	"testing"

	. "github.com/goghcrow/go-co"
)

func iter2slice[V any](g Iter[V]) (xs []V) {
	for i := range g {
		xs = append(xs, i)
	}
	return xs
}

func assertEqual(t *testing.T, got, expect any) {
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("expect\n%+v\n\ngot\n%+v", expect, got)
	}
}
