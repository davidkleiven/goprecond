package precond

import (
	"github.com/james-bowman/sparse"
	"gonum.org/v1/gonum/mat"
)

type CSRMulVecToer struct {
	Matrix  *sparse.CSR
	MatrixT *sparse.CSR
}

func NewCSRMulVecToer(matrix *sparse.CSR) CSRMulVecToer {
	return CSRMulVecToer{
		Matrix:  matrix,
		MatrixT: transposeCSR(matrix),
	}
}

func (c *CSRMulVecToer) MulVecTo(dst *mat.VecDense, trans bool, x mat.Vector) {
	n, _ := c.Matrix.Dims()

	mat := c.Matrix
	if trans {
		mat = c.MatrixT
	}

	for row := 0; row < n; row++ {
		sum := 0.0
		mat.DoRowNonZero(row, func(i, j int, v float64) {
			sum += v * x.AtVec(j)
		})
		dst.SetVec(row, sum)
	}
}
