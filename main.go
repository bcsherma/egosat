package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/bcsherma/egosat/egosat"
)

// parseFormula will construct a new solver
func parseFormula(filename string) (solver *egosat.Solver) {
	reader := createFormulaReader(filename)
	nVars, nClauses := parseProblemLine(reader)
	solver = egosat.CreateSolver(nClauses, nVars)
	parseClauses(reader, solver)
	return
}

// parseClauses will parse clauses from the formula in the reader and add them to
// the given solver
func parseClauses(reader *bufio.Reader, solver *egosat.Solver) {
	for {
		// TODO: This requires a newline at the end of the file, should accept
		// formulae without a newline at the end
		nextLine, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		var litVal int
		newClause := make([]egosat.Lit, 0, 10) // initialize capacity for 10
		for _, lit := range strings.Fields(nextLine) {
			litVal, err = strconv.Atoi(lit)
			if litVal == 0 {
				break
			}
			newClause = append(newClause, egosat.Lit(litVal))
		}
		if ok, _ := solver.AddClause(newClause, false); !ok {
			panic("Formula is trivially unsatisfiable!")
		}
	}
}

// createFormulaReader will open a reader for the formula with the comment
// lines already removed
func createFormulaReader(filename string) (reader *bufio.Reader) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	reader = bufio.NewReader(f)
	chompComments(reader)
	return
}

// chompComments takes as input a buffered file reader and consumes a prefix of
// lines beginning with the character c.
func chompComments(reader *bufio.Reader) {
	for {
		nextChar, err := reader.Peek(1)
		if err != nil {
			panic(err)
		}
		if nextChar[0] == 'c' {
			_, err := reader.ReadBytes('\n')
			if err != nil {
				panic(err)
			}
		} else {
			break
		}
	}
}

// parseProblemLine reads the problem line from a buffered file reader
// starting with the problem line.
func parseProblemLine(reader *bufio.Reader) (nVars int, nClauses int) {
	var err error
	problemLine, _ := reader.ReadString('\n')
	if problemLine[0] != 'p' {
		panic(fmt.Errorf("First char of problem line should be p, not %b", problemLine[0]))
	}
	fields := strings.Fields(string(problemLine))
	if fields[1] != "cnf" {
		panic(fmt.Errorf("Only cnf format is supported, received : %s ", fields[1]))
	}
	nVars, err = strconv.Atoi(fields[2])
	if err != nil {
		panic(fmt.Errorf("nbvars not an int : %q", fields[2]))
	}
	nClauses, err = strconv.Atoi(fields[3])
	if err != nil {
		panic(fmt.Errorf("nbClauses not an int : '%s'", fields[3]))
	}
	return
}

func main() {
	solver := parseFormula(os.Args[1])
	params := egosat.SolverParams{
		MaxConflict:         1000000,
		MaxLearnts:          solver.NumClauses() / 3,
		VarActivityDecay:    0.8,
		ClauseActivityDecay: 0.95,
	}
	for {
		res := solver.Search(params)
		if res == egosat.LFALSE {
			fmt.Println("s UNSATISFIABLE")
			break
		} else if res == egosat.LTRUE {
			fmt.Println("s SATISFIABLE")
			solver.PrintModel()
			break
		}
		params.MaxConflict = int(float32(params.MaxConflict) * 1.1)
		params.MaxLearnts = int(float32(params.MaxLearnts) * 1.5)
	}
	solver.PrintStats()
}
