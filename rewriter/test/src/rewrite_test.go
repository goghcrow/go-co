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

func ifIsNotTheLastInBlock_EmptyBranch() Iter[int] {
	if true {
	} else {
		Yield(1)
	}
	return nil
}

func ifIsNotTheLastInBlock_TwoBranchesReturn() Iter[int] {
	if true {
		Yield(1)
	} else {
		Yield(2)
	}
	return nil
}

func ifIsNotTheLastInBlock_MissingBranch1() Iter[int] {
	if true {
		Yield(1)
	}
	return nil
}

func ifIsNotTheLastInBlock_MissingBranch2() Iter[int] {
	if true {
		Yield(0)
		return nil
	}
	return nil
}

func ifIsNotTheLastInBlock_MissingBranch3() Iter[int] {
	if true {
		Yield(1)
	} else if true {
		Yield(2)
	}
	return nil
}

func ifIsNotTheLastInBlock_AllBranchesReturn() Iter[int] {
	if true {
		Yield(1)
	} else if true {
		Yield(2)
	} else {
		Yield(3)
	}
	return nil
}

func ifIsNotTheLastInBlock_AllBranchesReturn1() Iter[int] {
	if true {
		Yield(1)
	} else {
		if true {
			Yield(2)
		} else {
			Yield(3)
		}
	}
	return nil
}

func ifIsTheLastInBlock_EmptyBranch1() Iter[int] {
	for {
		if true {
		} else {
			Yield(1)
		}
	}
	return nil
}

func ifIsTheLastInBlock_EmptyBranch2() Iter[int] {
	for {
		if true {
		} else {
			Yield(1)
		}
	}
}

func ifIsTheLastInBlock_TwoBranchesReturn1() Iter[int] {
	for {
		if true {
			Yield(1)
		} else {
			Yield(2)
		}
	}
	return nil
}

func ifIsTheLastInBlock_TwoBranchesReturn2() Iter[int] {
	for {
		if true {
			Yield(1)
		} else {
			Yield(2)
		}
	}
}

func ifIsTheLastInBlock_MissingBranch11() Iter[int] {
	for {
		if true {
			Yield(1)
		}
	}
	return nil
}

func ifIsTheLastInBlock_MissingBranch12() Iter[int] {
	for {
		if true {
			Yield(1)
		}
	}
}

func ifIsTheLastInBlock_MissingBranch21() Iter[int] {
	for {
		if true {
			Yield(1)
		} else if true {
			Yield(2)
		}
	}
	return nil
}

func ifIsTheLastInBlock_MissingBranch22() Iter[int] {
	for {
		if true {
			Yield(1)
		} else if true {
			Yield(2)
		}
	}
}

func ifIsTheLastInBlock_AllBranchesReturn1() Iter[int] {
	for {
		if true {
			Yield(1)
		} else if true {
			Yield(2)
		} else {
			Yield(3)
		}
	}
	return nil
}

func ifIsTheLastInBlock_AllBranchesReturn2() Iter[int] {
	for {
		if true {
			Yield(1)
		} else if true {
			Yield(2)
		} else {
			Yield(3)
		}
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

// ↑ is equivalent to ↓
func block1() Iter[int] {
	for {
		Yield(1)
		break
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

// ↑ is equivalent to ↓
func block11() Iter[int] {
	for {
		for {
			Yield(1)
			break
		}
		break
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

// ↑ is equivalent to ↓
func block111() Iter[int] {
	for {
		i := 1
		for {
			i := 2
			Yield(1)
			println(i)
			break
		}
		println(i)
	}
	return nil
}

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
