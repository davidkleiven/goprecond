package property

import (
	"math"

	"gonum.org/v1/gonum/mat"
	"pgregory.net/rapid"
)

func DenseSquareMatrix(t *rapid.T, minSize int, maxSize int) *mat.Dense {
	n := rapid.IntRange(minSize, maxSize).Draw(t, "size")
	data := rapid.SliceOfN(rapid.Float64Range(-100.0, 100.0), n*n, n*n).Draw(t, "data")
	return mat.NewDense(n, n, data)
}

func SparseSymmetricMatrix(t *rapid.T, minSize int, maxSize int) *mat.SymDense {
	size := rapid.IntRange(minSize, maxSize).Draw(t, "size")
	matrix := mat.NewSymDense(size, nil)
	length := size * size / 3
	nonZero := rapid.SliceOfNDistinct(rapid.IntRange(0, size*size-1), length, length, func(item int) int { return item }).Draw(t, "non-zero")
	values := rapid.SliceOfN(rapid.Float64Range(-100.0, 100.0), length, length).Draw(t, "non-zero-values")

	for i := range nonZero {
		row := i % size
		col := i / size
		matrix.SetSym(row, col, values[i])
	}

	// Fill diagonal in a way that guarantees that the matrix is positive definite
	for i := 0; i < size; i++ {
		sum := 0.0
		for j := 0; j < size; j++ {
			sum += math.Abs(matrix.At(i, j))
		}
		matrix.SetSym(i, i, sum+0.1)
	}
	return matrix
}
