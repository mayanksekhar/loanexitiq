package loancalc

// ChartBar holds the geometry for one EMI's stacked bar in the dissection chart.
type ChartBar struct {
	X          float64
	Width      float64
	InterestY  float64
	InterestH  float64
	PrincipalY float64
	PrincipalH float64
	Dimmed     bool
}

// ChartData is everything the template needs to draw the EMI dissection chart.
type ChartData struct {
	Bars               []ChartBar
	MarkerX            float64
	MarkerLabelX       float64
	ExitMonth          int
	Width              float64
	Height             float64
	IntPctAtExit       int
	TotalPaid          float64
	InterestPaid       float64
	PrincipalCleared   float64
	Outstanding        float64
	LoanAmount         float64
	PrincipalPctOfLoan int
}

// BuildChart lays out one bar per EMI over a fixed viewBox.
func BuildChart(P, rate float64, tenure, exit int) ChartData {
	const w, h, pad = 1100.0, 300.0, 6.0

	emiVal, rows := Schedule(P, rate, tenure)
	slot := (w - pad*2) / float64(tenure)
	barW := slot - 1.2
	if barW < 1 {
		barW = 1
	}

	bars := make([]ChartBar, 0, tenure)
	var totalInt, totalPaid float64

	for i, r := range rows {
		x := pad + float64(i)*slot
		hI := (r.Interest / emiVal) * (h - 30)
		hP := (r.Principal / emiVal) * (h - 30)

		bars = append(bars, ChartBar{
			X:          x,
			Width:      barW,
			PrincipalY: h - hI - hP - 10,
			PrincipalH: hP,
			InterestY:  h - hI - 10,
			InterestH:  hI,
			Dimmed:     i >= exit,
		})

		if i < exit {
			totalInt += r.Interest
			totalPaid += r.Interest + r.Principal
		}
	}

	markerX := pad + float64(exit-1)*slot + barW/2
	labelX := markerX + 8
	if labelX > w-140 {
		labelX = w - 140
	}

	pct := 0
	if totalPaid > 0 {
		pct = int((totalInt/totalPaid)*100 + 0.5)
	}

	var outstanding float64
	if exit-1 < len(rows) {
		outstanding = rows[exit-1].Balance
	}
	principalCleared := P - outstanding
	prinPct := 0
	if P > 0 {
		prinPct = int(principalCleared / P * 100)
	}

	return ChartData{
		Bars:               bars,
		MarkerX:            markerX,
		MarkerLabelX:       labelX,
		ExitMonth:          exit,
		Width:              w,
		Height:             h,
		IntPctAtExit:       pct,
		TotalPaid:          totalPaid,
		InterestPaid:       totalInt,
		PrincipalCleared:   principalCleared,
		Outstanding:        outstanding,
		LoanAmount:         P,
		PrincipalPctOfLoan: prinPct,
	}
}
