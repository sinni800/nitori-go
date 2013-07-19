package main

import (
//"github.com/robertkrimen/otto"
)

type a struct {
	B string
	C func() string
}

func testit() {
	val, err := js.ToValue(a{"h", func() string { return "h" }})
	if err != nil {
		println(err.Error())
	} else {
		_, err := js.Run(`
		
			function lol(input) {
				console.log(input.C())
			}
		`)

		if err != nil {
			println(err.Error())
		} else {
			_, err := js.Call("lol", nil, val)
			if err != nil {
				println(err.Error())
			}
		}
	}
}
