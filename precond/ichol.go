package precond

import (
	"math"

	"github.com/james-bowman/sparse"
)

// Calculate the incomplete cholesky transformation
// The incomplete cholesky decompositoin is a special case of the
// incomplete LU transformation where U = L^T. Therefore, an instance
// if the ILUPreconditioner is returned
// The method panics if the provided matrix is not square or if the diagonal
// contain any non-positive elements
func IChol(A ZeroAwareMatrix) ILUPreconditioner {
	r, c := A.Dims()
	if r != c {
		panic("Matrix must be square")
	}

	lower := emptyDOK(r)
	A.DoNonZero(func(i, j int, v float64) { lower[i][j] = v })
	checkDiag(lower, r, true)

	for i := 0; i < r; i++ {
		diag := math.Sqrt(lower[i][i])
		lower[i][i] = diag
		for j := range lower[i] {
			if j > i {
				lower[j][i] /= diag
				for k := range lower[j] {
					if k > i {
						lower[j][k] -= lower[j][i] * lower[i][k]
					}
				}
			}
		}
	}

	lowerSpDOK := sparse.NewDOK(r, r)
	for i, row := range lower {
		for j, v := range row {
			lowerSpDOK.Set(i, j, v)
		}
	}

	lowerCSR := lowerSpDOK.ToCSR()
	upper := transposeCSR(lowerCSR)
	return ILUPreconditioner{
		lower:  lowerSpDOK.ToCSR(),
		upper:  upper,
		lowerT: upper,
		upperT: lowerCSR,
	}
}
