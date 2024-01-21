package src

import (
	. "github.com/goghcrow/go-co"
)

func ignoreReturn() (_ Iter[int]) {
	if true {
		var a Iter[int]
		return a
	} else if true {
		return nil
	} else {
		return
	}
}

func returnNil() Iter[int] {
	return nil
}

func returnNilInFuncLit() Iter[int] {
	f := func() any {
		return nil
	}
	f()
	Yield(0)
	return nil
}

func endlessLoop() Iter[int] {
	for {
	}
}

func endlessLoopWithReturn() Iter[int] {
	for {
	}
	return nil
}

func nestedEndlessLoopWithReturn() Iter[int] {
	for {
		for {

		}
	}
	return nil
}

func nestedForWith() Iter[int] {
	for {
		for {
			Yield(1)
			println(1)
		}
		println(2)
	}
	return nil
}

func nestedForIf() Iter[int] {
	for {
		if true {
			Yield(1)
			println(1)
		}
		println(2)
	}
	return nil
}

func endlessForWithReturn() Iter[int] {
	for {
		return nil
	}
}

func endlessForWithBreak() Iter[int] {
	for {
		break
	}
	return nil
}

func endlessForWithYieldThenBreak() Iter[int] {
	for {
		Yield(1)
		break
	}
	return nil
}

func endlessForWithYieldThenContinue() Iter[int] {
	for {
		Yield(1)
		continue
	}
}

func emptyBranchIfWithReturnNil() Iter[int] {
	if true {

	}
	return nil
}

func returnNilBranchIf() Iter[int] {
	if true {
		return nil
	}
	return nil
}

func endlessForWithYield() Iter[int] {
	for {
		Yield(1)
	}
}

func endlessForWithYieldThenReturn() Iter[int] {
	for {
		Yield(0)
		return nil
	}
}

func endlessForWithContinueBreakYield() Iter[int] {
	for {
		if true {
			continue
		} else if true {
			break
		} else {
			Yield(0)
		}
	}
	return nil
}

func endlessForWithYieldBreak() Iter[int] {
	for {
		if false {
			Yield(0)
		} else {
			break
		}
	}
	return nil
}

func endlessForWithBreakYield() Iter[int] {
	for {
		if false {
			break
		} else {
			Yield(0)
		}
	}
	return nil
}

func forWithContinueBreakYield() Iter[int] {
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			continue
		} else if i > 6 {
			break
		} else {
			Yield(i)
		}
	}
	return nil
}

func deadcode() Iter[int] {
	for {
		Yield(1)
	}
	return nil

	for {
		Yield(1)
	}
}

func block0() Iter[int] {
	{
		Yield(1)
	}
	return nil
}

func block00() Iter[int] {
	{
		{
			Yield(1)
		}
	}
	return nil
}

func block011() Iter[int] {
	for {
		i := 1
		{
			i := 2
			Yield(1)
			println(i)
		}
		println(i)
	}
	return nil
}

func breakContinue() Iter[int] {
	for {
		func() Iter[int] {
			Yield(1)
			func() {
				for {
					break
				}
			}()
			return nil
		}()
		continue
	}
	return nil
}
