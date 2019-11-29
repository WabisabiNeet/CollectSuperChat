package main

import (
	"fmt"
)

func main() {
	var a *A = nil
	defer fmt.Println("Hello World.")
	a.FuncA()
}

type A struct {
	Str string
}

func (a *A) FuncA() {
	go fmt.Println("FuncA" + a.Str)
}
