package main

import "fmt"

type greeting string

func (g greeting) Greet() {
	fmt.Println("Hejsan Världen!")
}

// exported
var Greeter greeting
