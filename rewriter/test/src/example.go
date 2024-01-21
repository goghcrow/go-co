package src

// func for0() Iter[int] {
// 	i := 0
// 	for Yield(1); i < 5; Yield(2) {
// 		Yield(3)
// 		i++
// 	}
// 	return nil
// }
//
// func for0_() Iter[int] {
// 	i := 0
// 	Yield(1)
// 	for i < 5 {
// 		Yield(3)
// 		i++
// 		Yield(2)
// 	}
// 	return nil
// }
//
// // 这个生成的 for 循环大概率后面缺 return
// func for1() Iter[int] {
// 	for Yield(1); ; {
//
// 	}
// }
//
// func for11() Iter[int] {
// 	for Yield(1); ; {
//
// 	}
// 	Yield(2)
// 	return nil
// }
//
// func for2() Iter[int] {
// 	for ; ; Yield(1) {
//
// 	}
// }
//
// func for22() Iter[int] {
// 	for ; ; Yield(1) {
// 		Yield(2)
// 	}
// }
//
// func for3() Iter[int] {
// 	for Yield(1); ; Yield(2) {
//
// 	}
// }
//
// func for4() Iter[int] {
// 	for Yield(1); ; Yield(2) {
// 		Yield(3)
// 	}
// }
//
// func for41() Iter[int] {
// 	i := 0
// 	for Yield(1); i < 5; Yield(2) {
// 		Yield(3)
// 		i++
// 	}
// 	return nil
// }
//
// func for5() Iter[int] {
// 	i := 0
// 	for Yield(1); ; Yield(2) {
// 		i++
// 		if true {
// 			i++
// 			Yield(3)
// 		}
// 		i++
// 	}
// 	println(i)
// 	return nil
// }
//
// func for6() Iter[int] {
// 	i := 0
// 	for Yield(1); ; Yield(2) {
// 		i++
// 		for {
// 			i++
// 			if true {
// 				i++
// 				Yield(3)
// 				i++
// 			}
// 			i++
// 		}
// 		i++
// 	}
// 	println(i)
// 	return nil
// }

// func mkCoroutine() Iter[int] {
// 	Yield(1)
// 	println(1)
//
// 	if false {
// 		Yield(2)
// 		println(2)
// 	} else {
// 		for i := 0; i < 10; i++ {
// 			Yield(i)
// 			println(i)
// 		}
// 	}
//
// 	func() Iter[int] {
// 		Yield(1)
// 		return nil
// 	}()
//
// 	return nil
// }
//
// func main() {
// 	for i := range mkCoroutine() {
// 		println(i)
// 	}
// }
//
// func xxx() Iter[int] {
// 	for i := range mkCoroutine() {
// 		Yield(i)
// 	}
// 	return nil
// }
//
// func useCoroutine() {
// 	{
// 		var i int
// 		for i = range mkCoroutine() {
// 			println(i)
// 		}
// 	}
// 	{
// 		for i := range mkCoroutine() {
// 			println(i)
// 		}
// 	}
// }

// func funcDecl() {
// 	println("funcDecl")
//
// 	if true {
// 		println("if")
// 	} else if true {
// 		println("else if")
// 	} else {
// 		println("else")
// 	}
//
// 	switch true {
// 	case true:
// 		println("switch")
// 	}
// 	var a fmt.Formatter
// 	switch a.(type) {
// 	default:
// 		println("type switch")
// 	}
//
// 	select {
// 	case <-make(chan int):
// 		println("select")
// 	}
//
// 	for {
// 		println("for")
// 		break
// 	}
//
// 	for _ = range []int{} {
// 		println("range")
// 	}
//
// 	var _ = func() {
// 		println("funcLit")
// 	}
//
// 	{
// 		println("block1")
// 		{
// 			println("block2")
// 		}
// 	}
// 	// {{}}
// }
//
// func name() co.Iter[int] {
// 	co.Yield(1)
// 	return nil
// }
