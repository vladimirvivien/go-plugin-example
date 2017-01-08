package main

import "C" // required

import "fmt"

type greeting string

func (g greeting) Greet() {
	fmt.Println("Hello Universe")
}

// exported
var Greeter greeting
