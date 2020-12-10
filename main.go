package main

import (
	"fmt"

	flag "github.com/spf13/pflag"
)

func main() {
	// -- Init
	var token string

	// -- Parse flags
	flag.StringVarP(&token, "token", "t", "", "the telegram token")
	flag.Parse()

	// TODO
	fmt.Println("token:", token)
}
