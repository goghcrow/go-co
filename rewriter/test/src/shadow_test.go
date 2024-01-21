package src

import (
	. "github.com/goghcrow/go-co"
)

func shadow0() Iter[int] {
	i := 0
	for i := 0; i < 2; i++ {
		Yield(i)
	}
	i++
	println(i)
	return nil
}

// ↑ is equivalent to ↓
func shadow1() Iter[int] {
	i := 0
	{
		i := 0
		for ; i < 2; i++ {
			Yield(i)
		}
	}
	i++
	println(i)
	return nil
}
