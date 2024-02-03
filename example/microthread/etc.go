package microthread

func assert(b bool) {
	if !b {
		panic(nil)
	}
}
