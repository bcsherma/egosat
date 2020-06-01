package main

import (
	"fmt"

	"github.com/bcsherma/egosat/egosat"
)

func main() {
	solver := egosat.CreateSolver("data/basic.cnf")
	result := solver.Search()
	if result == egosat.LFALSE {
		fmt.Println("UNSATISFIABLE")
	} else {
		fmt.Println("SATISFIABLE")
		solver.PrintModel()
	}
}
