package loancalc

import "math"

type Row struct {
	Month     int
	Interest  float64
	Principal float64
	Balance   float64
}

// EMI returns the monthly installment for principal P, annual rate (percent), over n months.
func EMI(P, annualRate float64, n int) float64 {
	r := annualRate / 1200
	pow := math.Pow(1+r, float64(n))
	return P * r * pow / (pow - 1)
}

// Schedule returns the full amortization schedule.
func Schedule(P, annualRate float64, n int) (float64, []Row) {
	r := annualRate / 1200
	e := EMI(P, annualRate, n)
	bal := P
	rows := make([]Row, 0, n)
	for k := 1; k <= n; k++ {
		i := bal * r
		p := e - i
		bal -= p
		if bal < 0 {
			bal = 0
		}
		rows = append(rows, Row{Month: k, Interest: i, Principal: p, Balance: bal})
	}
	return e, rows
}
