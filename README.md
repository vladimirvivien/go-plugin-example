# Go Plugin Example

This is a simple example that shows how to use the newly added `plugin` package in Go 1.8 (see https://tip.golang.org/pkg/plugin/).  A Go plugin is package compiled with the `-buildmode=plugin` which creates a shared library (`.so`) file.  As you will see here, using package plugin, Go can dynamically load the shared library at runtime to access exported functions an variables.

## Requirements
- Go 1.8 

## Restrictions
As of this writing, Go 1.8 (beta 2) only supports plugin on Linux OSes.  If you attempt to run this on OSX you will get `-buildmode=plugin not supported on darwin/amd64`.  This may change by the time GA comes out and may not be an issue.

## Pluggable Greeting System
To demo the plugin system, the example in this repository implements a simple greeting system.  Each plugin package (directory `./eng`, `./chi`) implements a greeting meesage in a different lanaguage.  File `greeter.go` uses the Go `plugin` package to load the plugin module and displays the proper message based on passed command-line parameters.

### The Plugin
To create a pluggable package is simple.  There are two requirements (so far):
- The package must be identified as `main`
- The package must `import "C"`

Let's examine the plugin in package directory `./eng`.  This plugin is responsible for displaying a message in english.  To do this it exports variable `Greeter` of local type `greeting` with method `Greet`.  

File [./eng/greeter.go](./eng/greeter.go)
```
package main

import "C" // required

import "fmt"

type greeting string

func (g greeting) Greet() {
	fmt.Println("Hello Universe")
}

// exported
var Greeter greeting
```
Notice a few things about the plugin module:
- Pluggable packages are basically regular Go packages
- The code in the package must `import "C"` as a requirement
- The exported variables and functions can be of any type (no documented restrictions I found)

The plugin package in directory `./chi` is exactly the same code except the message is in Chinese.

### Compiling the Plugins
The plugin package is compiled using the normal Go toolchain.  The only requirement is to use the `buildmode=plugin` compilation flag.  For our example, we will compile each shared library separately as shown below:
```
cd ./eng
go build -buildmode=plugin .
cd ../chi
go build -buildmode=plugin .
```
The compilation step will create `./eng/eng.so` and `./chi/chi.so` shared library files respectively.

### Using the Plugins
Once the plugin packages are compiled, they can be loaded dynamically using the built-in `plugin` package.  Let us examine file [./greeter.go](./greeter.go) to see how that is done. Loading and using a pluggable shared library is done in several steps.

#### Import the built-in plugin package

```
import (
	...
	"plugin"
)
```
#### Define/select type for imported elements (optional)
Remember, the exported elements, from the pluggable package, can be of any type.  For clarity purpose, in this example we defined interface type `Greeter` as a type that will be asserted against the exported element from the plugin. 
```
type Greeter interface {
	Greet()
}
```
#### Determine the .so file
The `.so` file must be in a location accessible from you program in order to open it.  In this example, the file .so file name is determined based on command-line argument and assigned to variable `mod`.
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
#### Open the plugin package
Using the standard library's `plugin` package, we can now open the plugin module.  That step creates a `*Plugin` variable as shown below.

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
#### Lookup Symbol
Next, we use the `*Plugin` to search for symbols that matches the name of the exported elements from the plugin module.  In our example plugin ([./eng/greeter.go](./eng/greeter.go), see earlier), we exported a variable called `Greeter`.  Therefore, we use `plug.Lookup("Greeter")` to locate that symbol.  The loaded symbol is then assigned to variable `symGreeter`.
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

#### Assert and use plugin value
Once we have the symbol loaded, we still have one additional step before we can use it.  We must use type assertion to validate that the symbol is of an expected type and assign its value to a variable of that type (well, the assignment step is optional).  In this example, we assert symbol `symGreeter` to be of interface type `Greeter` with `symGreeter.(Greeter)`.  Since the exported symbol from the plugin module `./eng/eng.so` is a variable with method `Greet` attached, the assertion is true and the value is assigned to variable `greeter`.  Lastly, we invoke the method from the plugin module with `greeter.Greet()`.
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
## Running the program
So now, we can display a message in english by running the program as:
```
> go run greeter.go english
Hello Universe
```
Or to do it in Chinese:
```
> go run greeter.go chinese
你好宇宙
```
The capability of the program is extended by the plugin allowing it to display a greeting message in different language without the need to recompile the program.
