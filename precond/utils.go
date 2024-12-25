package precond

import (
	"github.com/james-bowman/sparse"
)

func transposeCSR(matrix *sparse.CSR) *sparse.CSR {
	r, c := matrix.Dims()
	dok := sparse.NewDOK(r, c)
	matrix.DoNonZero(func(i, j int, v float64) {
		dok.Set(j, i, v)
	})
	return dok.ToCSR()
}
