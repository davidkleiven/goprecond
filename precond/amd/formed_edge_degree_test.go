package amd

import (
	"math/rand"
	"slices"
	"testing"
	"testing/quick"

	"github.com/davidkleiven/goprecond/precond/property"
)

func TestDegree(t *testing.T) {
	for _, test := range []struct {
		adjList [][]int
		remove  int
		want    int
		desc    string
	}{
		{
			adjList: [][]int{{1}, {0, 2}, {1}},
			remove:  0,
			want:    1,
			desc:    "Remove edge node, no new edges are formed",
		},
		{
			adjList: [][]int{{1}, {0, 2}, {1}},
			remove:  1,
			want:    3,
			desc:    "Two nearest neighbours, and one new edge 0 - 2 is formed",
		},
		{
			adjList: [][]int{{1, 2}, {0, 2}, {1, 2}},
			remove:  1,
			want:    2,
			desc:    "Two nearest neighbours, but edge 0-2 already exists",
		},
	} {
		degreeCalc := NewFormedEdgeDegree(test.adjList)
		ctx := NewAmdCtx(len(test.adjList))
		degree := degreeCalc.Degree(test.remove, test.adjList, &ctx)

		if degree != test.want {
			t.Errorf("%s:\nwanted %d got %d", test.desc, test.want, degree)
		}
	}
}

func TestUpdateDegree(t *testing.T) {
	for _, test := range []struct {
		adjList    [][]int
		remove     int
		want       [][]int
		eliminated []int
		desc       string
	}{
		{
			adjList:    [][]int{{1}, {0, 2}, {1}},
			remove:     0,
			want:       [][]int{{1}, {0, 2}, {1}},
			eliminated: []int{},
			desc:       "Remove edge node. No new edges",
		},
		{
			adjList:    [][]int{{1}, {0, 2}, {1}},
			remove:     1,
			want:       [][]int{{1, 2}, {0, 2}, {1, 0}},
			eliminated: []int{},
			desc:       "Remove center node. New edge between 0 and 2",
		},
		{
			adjList:    [][]int{{1}, {0, 2}, {1}},
			remove:     1,
			want:       [][]int{{1}, {0, 2}, {1}},
			eliminated: []int{2},
			desc:       "Remove center node. But node 2 is already eliminated. No new edge",
		},
	} {
		ctx := NewAmdCtx(len(test.adjList))
		for _, node := range test.eliminated {
			ctx.Eliminated[node] = true
		}

		degreeCalc := NewFormedEdgeDegree(test.adjList)

		eliminated := Node{index: test.remove}
		degreeCalc.OnNodeEliminated(eliminated, test.adjList, &ctx)

		for i := range test.adjList {
			if !slices.Equal(test.adjList[i], test.want[i]) {
				t.Errorf("Test %s: wanted %v\ngot\n%v\n", test.desc, test.want, test.adjList)
				break
			}
		}
	}
}

func TestEdgeUpdates(t *testing.T) {
	config := quick.Config{
		Rand:     rand.New(rand.NewSource(0)),
		MaxCount: 10,
	}

	validAdjList := func(gen property.SparseSymmetrixMatrixGenerator) bool {
		adjList := property.SparseMatToAdjList(gen.Matrix)
		degreeCalc := NewFormedEdgeDegree(adjList)
		ctx := NewAmdCtx(len(adjList))
		for i := 0; i < len(adjList); i++ {
			ctx.Eliminated[i] = true
			degreeCalc.OnNodeEliminated(Node{index: i}, adjList, &ctx)

			// On every step the adjecancy list should be valid (e.g. no duplicates)
			for i, neighbours := range adjList {
				uniqueNeighbours := make(map[int]bool)
				for _, neighbour := range neighbours {
					uniqueNeighbours[neighbour] = true

					// Check that i is among the neighbours
					oppositeRelationExists := false
					for _, oppositeNeighbour := range adjList[neighbour] {
						if oppositeNeighbour == i {
							oppositeRelationExists = true
							break
						}
					}

					if !oppositeRelationExists {
						return false
					}
				}
				if len(uniqueNeighbours) != len(neighbours) {
					return false
				}
			}
		}
		return true
	}

	if err := quick.Check(validAdjList, &config); err != nil {
		t.Error(err)
	}
}
