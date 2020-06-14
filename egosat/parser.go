package egosat

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// CreateSolver will construct a new solver
func CreateSolver(filename string) (solver *Solver) {
	reader := createFormulaReader(filename)
	nVars, nClauses := parseProblemLine(reader)
	solver = &Solver{
		clauses:          make([]*Clause, 0, nClauses),
		learntClauses:    make([]*Clause, 0, 100),
		watcherLists:     make([][]*Clause, 2*nVars),
		assignments:      make([]Lbool, nVars+1),
		trail:            make([]Lit, 0, nVars),
		reasons:          make([]*Clause, nVars+1),
		level:            make([]int, nVars+1),
		variableActivity: make([]float32, nVars+1),
	}
	solver.variableOrder = createQueue(solver, nVars)
	for i := 1; i <= nVars; i++ {
		solver.variableActivity[i] = 1e6
		solver.variableOrder.insert(i)
	}
	parseClauses(reader, solver)
	return
}

// parseClauses will parse clauses from the formula in the reader and add them to
// the given solver
func parseClauses(reader *bufio.Reader, solver *Solver) {
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
		newClause := make([]Lit, 0, 10) // initialize capacity for 10
		for _, lit := range strings.Fields(nextLine) {
			litVal, err = strconv.Atoi(lit)
			if litVal == 0 {
				break
			}
			newClause = append(newClause, Lit(litVal))
		}
		if ok, _ := solver.addClause(newClause, false); !ok {
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
