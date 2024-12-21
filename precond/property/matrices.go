package property

import (
	"math/rand"
	"reflect"

	"gonum.org/v1/gonum/mat"
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
