package egosat

import "testing"

func TestPriorityQueue(t *testing.T) {
	solver := &Solver{literalActivity: []float64{0.11, 1.32, 2.31, -0.123, 2.44, 3.1}}
	queue := createQueue(solver, 5)
	for i := 1; i < 4; i++ {
		queue.insert(Lit(i))
		queue.insert(Lit(-i))
	}
	if queue.removeMax() != Lit(-3) {
		t.Fail()
	}
	if queue.removeMax() != Lit(3) {
		t.Fail()
	}
	if queue.removeMax() != Lit(2) {
		t.Fail()
	}
}
