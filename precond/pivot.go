package precond

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

// Pivot represent a pivot matrix
// It implements mat.Matrix interface from gonum
type Pivot struct {
	Pivots []int
}

// NoPivot returns the identify pivot matrix
func NoPivot(N int) Pivot {
	data := make([]int, N)
	for i := 0; i < N; i++ {
		data[i] = i
	}
	return Pivot{data}
}

// Dims return the dimensions of the matrix
func (p *Pivot) Dims() (int, int) {
	return len(p.Pivots), len(p.Pivots)
}

// At return the (i, j) element of the pivot matrix
func (p *Pivot) At(i, j int) float64 {
	if p.Pivots[i] == j {
		return 1.0
	}
	return 0.0
}

// T performs implicit transpose of the pivot matrix
func (p *Pivot) T() mat.Matrix {
	return &mat.Transpose{Matrix: p}
}

// PartialPivotMatrix calculates a pivot matrix that re-orders the rows
// such that the diagonal is larger (in absolute value) than all elements
// below it
//
// Example:
// [0, 1,  2]         [6, 10, 8]
// [3, 7,  5]  ---->  [3,  7, 5]
// [6, 10, 8]         [0,  1, 2]
// The pivot matrix returned in the example would be
//
// P =
//
//	[0, 0, 1]
//	[0, 1, 0]
//	[1, 0, 0]
func PartialPivotMatrix(matrix mat.ColNonZeroDoer, N int) Pivot {
	pivot := make([]int, N)
	alreadyUsed := make(map[int]bool)

	for i := 0; i < N; i++ {
		maxRow := 0
		maxAbsValue := -1.0

		matrix.DoColNonZero(i, func(i, j int, v float64) {
			_, isUsed := alreadyUsed[i]
			if absV := math.Abs(v); absV > maxAbsValue && !isUsed {
				maxAbsValue = absV
				maxRow = i
			}
		})
		alreadyUsed[maxRow] = true
		pivot[i] = maxRow
	}
	return Pivot{Pivots: pivot}
}
