package amd

import "testing"

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
