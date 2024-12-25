package property

import (
	"gonum.org/v1/gonum/mat"
)

func SparseMatToAdjList(mat *mat.SymDense) [][]int {
	r := mat.SymmetricDim()
	adjList := make([][]int, r)
	for i := 0; i < r; i++ {
		adjList[i] = make([]int, 0)
		for j := 0; j < r; j++ {
			if mat.At(i, j) != 0.0 {
				adjList[i] = append(adjList[i], j)
			}
		}
	}
	return adjList
}
