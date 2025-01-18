package amd

import (
	"slices"
	"testing"
)

type testCase struct {
	degCalc *WeightedEnode
	adjList [][]int
	ctx     *AmdCtx
}

func triangleCase() testCase {
	weightedEnode := NewWeightedEnode(3)
	adjList := [][]int{
		{1, 2},
		{0, 2},
		{0, 1},
	}
	ctx := NewAmdCtx(3)
	return testCase{
		degCalc: weightedEnode,
		adjList: adjList,
		ctx:     &ctx,
	}
}
func TestWeightedEnodeDegree(t *testing.T) {
	tc := triangleCase()

	deg := tc.degCalc.Degree(0, tc.adjList, tc.ctx)
	if deg != 2 {
		t.Errorf("Wanted %d got %d\n", 2, deg)
	}

	tc.degCalc.weights[0] = 2
	deg = tc.degCalc.Degree(1, tc.adjList, tc.ctx)
	if deg != 3 {
		t.Errorf("Wanted %d got %d\n", 3, deg)
	}
}

func TestWeightedEnodeOnUpdateNode(t *testing.T) {
	tc := triangleCase()
	tc.degCalc.OnNodeEliminated(Node{index: 0, degree: 2}, tc.adjList, tc.ctx)
	expectedWeights := []int{1, 1, 1}
	if slices.Compare(expectedWeights, tc.degCalc.weights) != 0 {
		t.Errorf("Wanted %v got %v\n", expectedWeights, tc.degCalc.weights)
	}
}
