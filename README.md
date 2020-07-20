# egosat

### A SAT solver written in Go, based on `minisat`

## Installation

You can install egosat by running
```
go get github.com/bcsherma/egosat && go install github.com/bcsherma/egosat 
```

## Usage

For the moment, `egosat` has not been extended to handle common variations of
the SAT problem, e.g. max-SAT, handling XORs natively... The solver takes as
input a CNF file and either finds a satisfying assignment or determines that no
such assignment exists. The basic usage is

```
egosat my_formula.cnf
```

## Why?

I have always wanted to write a SAT solver since I first learned about the
problem and the annual satisfiability competiton. This is a vanity project
(hence the name) for me and for the time being this should only serve as a
well-documented, straightforward adaptation of `minisat`. Many of the best ideas
to come out of satisfiability testing research in the last few decades have yet
to be incoporated into `egosat`, e.g. variable reordering, clause subsumption...

If you are looking for a SAT solver to incorporate into a project, I would
suggest looking elsewhere. There are super-duper performant SAT solvers written
in C++ like `cryptominisat`. For Go users, `gini` and `gophersat` are both
excellent, French solvers written in go.