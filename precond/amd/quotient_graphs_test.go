package amd

import (
	"math/rand"
	"slices"
	"testing"
)

func TestQuotientGraph(t *testing.T) {
	// Build the graph
	//  0 - 1 - 2 - 4
	//  |  /    |
	//  3 -------
	adjList := [][]int{
		{1, 3},
		{0, 2, 3},
		{1, 3, 4},
		{0, 1, 2},
		{2},
	}

	quotientGraphCalc := QuotientGraphExactDegreeCalculator{}
	ctx := NewAmdCtx(4)

	for _, test := range []struct {
		eliminated []bool
		node       int
		want       int
		desc       string
	}{

		{
			eliminated: []bool{false, false, false, false, false},
			node:       0,
			want:       2,
			desc:       "Degree of node 0 is 2",
		},
		{
			eliminated: []bool{false, false, false, false, false},
			node:       3,
			want:       3,
			desc:       "Degree of node 3 is 3",
		},
		{
			eliminated: []bool{false, false, false, true, false},
			node:       0,
			want:       2,
			desc:       "Degree of node 0 is 2 because 0-2 is formed (but 0-3 no longer exist)",
		},
		{
			eliminated: []bool{false, true, true, false, false},
			node:       0,
			want:       2,
			desc:       "Degree of node 0 is 2 (the edges are 0-3 and 0-4)",
		},
	} {
		ctx.Eliminated = test.eliminated
		degree := quotientGraphCalc.Degree(test.node, adjList, &ctx)
		if degree != test.want {
			t.Errorf("Test %s: wanted %d got %d\n", test.desc, test.want, degree)
		}
	}
}

func TestOrderChain(t *testing.T) {
	n := 50
	rnd := rand.New(rand.NewSource(0))
	nodes := make([]int, n)
	for i := range nodes {
		nodes[i] = i
	}

	rnd.Shuffle(n, func(i, j int) {
		// Ensure that the largest node is an edge
		// In that case the minimum degree method
		// should be able to recover the order of a chain
		if i != n-1 && j != n-1 {
			nodes[i], nodes[j] = nodes[j], nodes[i]
		}
	})

	adjList := make([][]int, n)
	for i := range n {
		current := nodes[i]
		neighbours := make([]int, 0, 2)
		if i > 0 {
			neighbours = append(neighbours, nodes[i-1])
		}
		if i < n-1 {
			neighbours = append(neighbours, nodes[i+1])
		}
		adjList[current] = neighbours
	}

	quotientGraphCalc := QuotientGraphExactDegreeCalculator{}
	pivot := ApproximateMinimumDegree(n, adjList, &quotientGraphCalc)

	pivotedNodes := make([]int, n)
	for i, v := range pivot {
		pivotedNodes[v] = nodes[i]
	}

	slices.Sort(nodes)
	if slices.Compare(pivotedNodes, nodes) != 0 {
		t.Errorf("AMD should be able to recover the order of the 1D chain. Got\n%v\n", pivotedNodes)
	}
}
