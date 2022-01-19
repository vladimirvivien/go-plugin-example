package main

import "fmt"

type greeting string

func (g greeting) Greet() {
    fmt.Println("Здравствуй, вселенная!")
}

// exported
var Greeter greeting
