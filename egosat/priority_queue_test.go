package egosat

import "testing"

func TestPriorityQueue(t *testing.T) {
	solver := &Solver{variableActivity: []float32{0, 1.32, 2.31, -0.123, 2.44, 3.1}}
	queue := createQueue(solver, 5)
	queue.insert(1)
	if queue.heap[0] != 1 {
		t.Fail()
	}
	queue.insert(2)
	queue.insert(3)
	queue.insert(4)
	queue.insert(5)
	if queue.removeMin() != 3 {
		t.Fail()
	}
	if queue.removeMin() != 1 {
		t.Fail()
	}
	if queue.removeMin() != 2 {
		t.Fail()
	}
}
