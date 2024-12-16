package precond

import (
	"errors"
	"fmt"
	"math"

	"github.com/james-bowman/sparse"
	"gonum.org/v1/gonum/mat"
)

type ZeroAwareMatrix interface {
	mat.Matrix
	mat.NonZeroDoer
}

type ILUPreconditioner struct {
	lower *sparse.CSR
	upper *sparse.CSR
}

func (ilu *ILUPreconditioner) checkDimensions(dst *mat.VecDense, rhs mat.Vector) error {
	n, _ := ilu.lower.Dims()
	dstDim, _ := dst.Dims()
	rhsDim, _ := rhs.Dims()

	if dstDim != n || rhsDim != n {
		return fmt.Errorf("expected lengths to be %d, got dst: %d and rhs: %d", n, dstDim, rhsDim)
	}
	return nil
}

// SolveVecTo solves the linear system of equation given by
// Ax = b using the LU transformation stored in the receiver
func (ilu *ILUPreconditioner) SolveVecTo(dst *mat.VecDense, trans bool, rhs mat.Vector) error {
	if err := ilu.checkDimensions(dst, rhs); err != nil {
		return err
	}

	if trans {
		return errors.New("transpose is currently not supported")
	}
	n, _ := dst.Dims()
	tmpSolution := mat.NewVecDense(n, nil)
	ilu.forwardSubstition(tmpSolution, rhs)
	ilu.backwardSubstitiion(dst, tmpSolution)
	return nil
}

func (ilu *ILUPreconditioner) forwardSubstition(dst *mat.VecDense, rhs mat.Vector) {
	n, _ := rhs.Dims()
	for i := 0; i < n; i++ {
		sum := 0.0
		ilu.lower.DoRowNonZero(i, func(i, j int, v float64) {
			if j < i {
				sum += v * dst.AtVec(j)
			}
		})
		dst.SetVec(i, (rhs.AtVec(i)-sum)/ilu.lower.At(i, i))
	}
}

func (ilu *ILUPreconditioner) backwardSubstitiion(dst *mat.VecDense, rhs mat.Vector) {
	n, _ := rhs.Dims()
	for i := n - 1; i >= 0; i-- {
		sum := 0.0
		ilu.upper.DoRowNonZero(i, func(i, j int, v float64) {
			if j > i {
				sum += v * dst.AtVec(j)
			}
		})

		dst.SetVec(i, (rhs.AtVec(i)-sum)/ilu.upper.At(i, i))
	}
}

func checkDiag(lu map[int]map[int]float64, N int) {
	tol := 1e-8
	for i := 0; i < N; i++ {
		if diag, ok := lu[i][i]; !ok || math.Abs(diag) < tol {
			panic("Zero on diagonal")
		}
	}
}

func emptyDOK(N int) map[int]map[int]float64 {
	dok := make(map[int]map[int]float64)
	for i := 0; i < N; i++ {
		dok[i] = make(map[int]float64)
	}
	return dok
}

// ILUZero calculates the incomplete LU decomposition of the matrix A
// If A is dense, this is the same as the complete LU decomposition
// The method panics if A is not square or the diagonal contains zeros
func ILUZero(A ZeroAwareMatrix) ILUPreconditioner {
	nrows, ncols := A.Dims()
	if nrows != ncols {
		panic("Matrix must be square")
	}

	lu := emptyDOK(nrows)

	// Transfer matrix into temporary lu
	A.DoNonZero(func(i, j int, v float64) {
		lu[i][j] = v
	})
	checkDiag(lu, nrows)

	for i := 0; i < nrows; i++ {
		diag := lu[i][i]
		for j := range lu[i] {
			if j > i {
				lu[j][i] /= diag
				for k := range lu[j] {
					if k > i {
						lu[j][k] -= lu[j][i] * lu[i][k]
					}
				}
			}
		}
	}

	// Collect upper/lower matrices
	upper := sparse.NewDOK(nrows, nrows)
	lower := sparse.NewDOK(nrows, ncols)
	for i, row := range lu {
		for j, v := range row {
			if j >= i {
				upper.Set(i, j, v)
			} else {
				lower.Set(i, j, v)
			}
		}
	}

	for i := 0; i < nrows; i++ {
		lower.Set(i, i, 1.0)
	}
	return ILUPreconditioner{
		upper: upper.ToCSR(),
		lower: lower.ToCSR(),
	}
}
