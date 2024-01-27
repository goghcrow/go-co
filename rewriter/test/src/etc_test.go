package src

import (
	"reflect"
	"testing"

	. "github.com/goghcrow/go-co"
)

func iter2slice[V any](g Iter[V]) []V {
	var s []V
	for i := range g {
		s = append(s, i)
	}
	return s
}

func assertEqual(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Logf("expect %+v got %+v", b, a)
		t.FailNow()
	}
}
