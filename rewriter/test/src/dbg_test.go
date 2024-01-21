package src

import (
	. "github.com/goghcrow/go-co"
)

func ifElsIf2222() Iter[int] {
	Yield(1)

	return nil
}

//
// func breakContinue11111111() Iter[int] {
// 	func() (_ Iter[int]) {
// 		Yield(1)
// 		func() {
// 			for {
// 				break
// 			}
// 		}()
// 		if false {
// 			return
// 		}
// 		return nil
// 	}()
// 	return nil
// }

//
// func TestReturnBug___(t *testing.T) {
// 	g := func() Iter[int] {
// 		i := 0
// 		for ; i < 10; i++ {
// 			if false {
// 				break
// 			}
// 		}
// 		Yield(i)
// 		return nil
// 	}
// 	xs := iter2slice(g())
// 	assertEqual(t, xs, []int{10})
// }
//
// func TestReturnBug______(t *testing.T) {
// 	g := func() Iter[int] {
// 		i := 0
// 		for ; i < 10; i++ {
// 			{
// 				if false {
// 					break
// 				}
// 			}
// 		}
// 		Yield(i)
// 		return nil
// 	}
// 	xs := iter2slice(g())
// 	assertEqual(t, xs, []int{10})
// }

// func TestTrivalBlock111(t *testing.T) {
// 	g := func() Iter[int] {
// 		for {
// 			{
// 				if true {
// 					break
// 				} else {
// 					continue
// 				}
// 			}
// 		}
// 		Yield(1)
// 		return nil
// 	}
// 	xs := iter2slice(g())
// 	assertEqual(t, xs, []int{1})
// }
