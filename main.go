package main

import (
	"fmt"
	"os"
	
	"orb/parser"
)


func main() {
	tree, err := parser.Parse(os.Stdin)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(tree)
}
