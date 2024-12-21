package precond

import "gonum.org/v1/gonum/mat"

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
