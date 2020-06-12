package main

import (
	"fmt"
	"os"

	"github.com/bcsherma/egosat/egosat"
)

func main() {
	solver := egosat.CreateSolver(os.Args[1])
	result := solver.Search()
	if result == egosat.LFALSE {
		fmt.Println("s UNSATISFIABLE")
	} else {
		fmt.Println("s SATISFIABLE")
		solver.PrintModel()
	}
}
