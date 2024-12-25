package precondtest

import "gonum.org/v1/gonum/mat"

type DenseNonZeroDoer struct {
	*mat.Dense
}

func (d *DenseNonZeroDoer) DoNonZero(fn func(i, j int, v float64)) {
	nrows, ncols := d.Dims()
	for row := 0; row < nrows; row++ {
		for col := 0; col < ncols; col++ {
			if v := d.At(row, col); v != 0.0 {
				fn(row, col, v)
			}
		}
	}
}

func (d *DenseNonZeroDoer) DoRowNonZero(row int, fn func(i, j int, v float64)) {
	_, ncols := d.Dims()
	for j := 0; j < ncols; j++ {
		if v := d.At(row, j); v != 0.0 {
			fn(row, j, v)
		}
	}
}

func (d *DenseNonZeroDoer) DoColNonZero(col int, fn func(i, j int, v float64)) {
	nrows, _ := d.Dims()
	for i := 0; i < nrows; i++ {
		if v := d.At(i, col); v != 0.0 {
			fn(i, col, v)
		}
	}
}
