package property

import (
	"math"
	"math/rand"
	"reflect"

	"gonum.org/v1/gonum/mat"
	"pgregory.net/rapid"
)

type DenseSquareMatrixGenerator struct {
	Matrix *mat.Dense
}

func (d DenseSquareMatrixGenerator) Generate(rand *rand.Rand, size int) reflect.Value {
	matrix := mat.NewDense(size, size, nil)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			matrix.Set(i, j, rand.NormFloat64())
		}
	}

	return reflect.ValueOf(DenseSquareMatrixGenerator{
		Matrix: matrix,
	})
}

func DenseSquareMatrix(t *rapid.T, minSize int, maxSize int) *mat.Dense {
	n := rapid.IntRange(minSize, maxSize).Draw(t, "size")
	data := rapid.SliceOfN(rapid.Float64Range(-100.0, 100.0), n*n, n*n).Draw(t, "data")
	return mat.NewDense(n, n, data)
}

type SparseSymmetrixMatrixGenerator struct {
	Matrix *mat.SymDense
}

func (d SparseSymmetrixMatrixGenerator) Generate(rand *rand.Rand, size int) reflect.Value {
	size = size % 100
	matrix := mat.NewSymDense(size, nil)

	for i := 0; i < size; i++ {
		for j := i + 1; j < size; j++ {
			if rand.Int()%3 == 0 {
				matrix.SetSym(i, j, rand.NormFloat64())
			}
		}
	}

	// Fill diagonal in a way that guarantees that the matrix is positive definite
	for i := 0; i < size; i++ {
		sum := 0.0
		for j := 0; j < size; j++ {
			sum += math.Abs(matrix.At(i, j))
		}
		matrix.SetSym(i, i, sum)
	}

	return reflect.ValueOf(SparseSymmetrixMatrixGenerator{
		Matrix: matrix,
	})
}
