package src

import (
	. "github.com/goghcrow/go-co"
)

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
