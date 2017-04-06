# Go Plugin Example

The code in this repository shows how to use the new `plugin` package in Go 1.8 (see https://tip.golang.org/pkg/plugin/).  A Go plugin is package compiled with the `-buildmode=plugin` which creates a shared object (`.so`) library file instead of the standar archive (`.a`) library file.  As you will see here, using the standar library's `plugin` package, Go can dynamically load the shared object file at runtime to access exported elements such as functions an variables.

You can read the related article [on Medium](https://medium.com/learning-the-go-programming-language/writing-modular-go-programs-with-plugins-ec46381ee1a9).

## Requirements
The plugin system requires Go version 1.8.  At this time, it is only supports plugin on Linux.  Attempt 
to compile plugins on OSX, for instance, will result in  `-buildmode=plugin not supported on darwin/amd64` error.

## A Pluggable Greeting System
The demo in this repository implements a simple greeting system.  Each plugin package (directories `./eng` and `./chi`) implements code that prints a greeting meesage in a different lanaguage.  File `./greeter.go` uses the new Go `plugin` package to load the pluggable modules and displays the proper message using passed command-line parameters.

For instance, when the program is executed it prints a greeting in English or Chinese 
using the passed parameter to select the plugin to load for the appropriate language.
```
> go run greeter.go english
Hello Universe
```
Or to do it in Chinese:
```
> go run greeter.go chinese
你好宇宙
```
As you can see, the capability of the driver program is dynamically expanded by the plugins allowing it to display a greeting message in different language without the need to recompile the program.

Let us see how this is done.


## The Plugin
To create a pluggable package is simple.  Simply create a regular Go package designated as `main`. Use the capitalization rule to indicate functions and variables that are exported as part of the plugin.  This is shown below in file  `./eng/greeter.go`.  This plugin is responsible for displaying a message in `English`.  

File [./eng/greeter.go](./eng/greeter.go)

```go
package main

import "fmt"

type greeting string

func (g greeting) Greet() {
	fmt.Println("Hello Universe")
}

// this is exported
var Greeter greeting
```
Notice a few things about the pluggable module:

- Pluggable packages are basically regular Go packages
- The package must be marked `main`
- The exported variables and functions can be of any type (I found no documented restrictions)

The previous code exports variable `Greeter` of type `greeting`.  As we will see later, the code that will consume this exported value must have a compatible type for assertion.  One way this can be handled is to have an interface with the same method set. (The plugin package in directory `./chi` is exactly the same code except the message is in Chinese.)

### Compiling the Plugins
The plugin package is compiled using the normal Go toolchain.  The only requirement is to use the `buildmode=plugin` compilation flag as shown below:

```
go build -buildmode=plugin -o eng/eng.so eng/greeter.go
go build -buildmode=plugin -o chi/chi.so chi/greeter.go
```
The compilation step will create `./eng/eng.so` and `./chi/chi.so` plugin files respectively.

### Using the Plugins
Once the plugin modules are available, they can be loaded dynamically using the Go standard library's `plugin` package.  Let us examine file [./greeter.go](./greeter.go), the driver program that loads and uses the plugin at runtime. Loading and using a shared object library is done in several steps as outlined below:

#### 1. Import package plugin
```
import (
	...
	"plugin"
)
```
#### 2. Define/select type for imported elements (optional)
The exported elements, from the pluggable package, can be of any type.  The consumer code, loading the plugin, must have a compatible type defined (or pre-defined in case of built-in types) for assertion. In this example we define interface type `Greeter` as a type that will be asserted against the exported variable from the plugin module. 
```
type Greeter interface {
	Greet()
}
```
#### 3. Determine the .so file to load
The `.so` file must be in a location accessible from you program in order to open it.  In this example, the file .so files are located in directories `./eng` and `./chi`.  and are selected based on the value of a command-line argument.  The selected name is then assigned to variable `mod`.
```
func main() {
	// determine module to load
	lang := "english"
	if len(os.Args) == 2 {
		lang = os.Args[1]
	}
	var mod string
	switch lang {
	case "english":
		mod = "./eng/eng.so"
	case "chinese":
		mod = "./chi/chi.so"
	default:
		fmt.Println("don't speak that language")
		os.Exit(1)
	}
...
```
#### 4. Open the plugin package
Using the Go standard library's `plugin` package, we can now `open` the plugin module.  That step creates a value of type `*plugin.Plugin`.  It is used later to manage access to the plugin's exported elements.

```
func main(){
...
	// load module
	// 1. open the so file to load the symbols
	plug, err := plugin.Open(mod)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
...
```
#### 6. Lookup a Symbol
Next, we use the `*plugin.Plugin` value to search for symbols that matches the name of the exported elements from the plugin module.  In our example plugin ([./eng/greeter.go](./eng/greeter.go), seen earlier), we exported a variable called `Greeter`.  Therefore, we use `plug.Lookup("Greeter")` to locate that symbol.  The loaded symbol is then assigned to variable `symGreeter` (of type `package.Symbol`).
```
func main(){
...
	// 2. look up a symbol (an exported function or variable)
	// in this case, variable Greeter
	symGreeter, err := plug.Lookup("Greeter")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
...
```

#### 7. Assert the symbol's type and use it
Once we have the symbol loaded, we still have one additional step before we can use it.  We must use type assertion to validate that the symbol is of an expected type and (optionally) assign its value to a variable of that type.  In this example, we assert symbol `symGreeter` to be of interface type `Greeter` with `symGreeter.(Greeter)`.  Since the exported symbol from the plugin module `./eng/eng.so` is a variable with method `Greet` attached, the assertion is true and the value is assigned to variable `greeter`.  Lastly, we invoke the method from the plugin module with `greeter.Greet()`.
```
func main(){
...
	// 3. Assert that loaded symbol is of a desired type
	// in this case interface type Greeter (defined above)
	var greeter Greeter
	greeter, ok := symGreeter.(Greeter)
	if !ok {
		fmt.Println("unexpected type from module symbol")
		os.Exit(1)
	}

	// 4. use the module
	greeter.Greet()

}
```
