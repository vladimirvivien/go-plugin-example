package main

import "C"

import "fmt"

type greeting string

func (g greeting) Greet() {
	fmt.Println("你好宇宙")
}

var Greeter greeting
