package main

import (
	"strconv"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/mayanksekhar/loanexitiq/go-service/loancalc"
	"github.com/mayanksekhar/loanexitiq/go-service/templates"
)

func clampExit(tenure, exit int) int {
	if exit > tenure-3 {
		exit = tenure - 3
	}
	if exit < 1 {
		exit = 1
	}
	return exit
}

func computeStats(amount, rate float64, tenure, exit int) templates.StatsData {
	exit = clampExit(tenure, exit)
	emi, rows := loancalc.Schedule(amount, rate, tenure)

	var interestPaid float64
	for i := 0; i < exit; i++ {
		interestPaid += rows[i].Interest
	}
	outstanding := rows[exit-1].Balance
	principalPaid := amount - outstanding

	return templates.StatsData{
		EMI:           emi,
		InterestPaid:  interestPaid,
		PrincipalPaid: principalPaid,
		Outstanding:   outstanding,
		ExitMonth:     exit,
		Tenure:        tenure,
	}
}

func computeBankCost(amount, rate float64, tenure, exit, bankIdx int) templates.LenderResult {
	exit = clampExit(tenure, exit)
	if bankIdx < 0 || bankIdx >= len(loancalc.Lenders) {
		bankIdx = 0
	}
	r := loancalc.ComputeForLender(amount, rate, tenure, exit, true, bankIdx)
	return templates.LenderResult{
		Name:         r.Name,
		Note:         r.Note,
		RatePct:      r.RatePct,
		InterestPaid: r.InterestPaid,
		Fee:          r.Fee,
		FeePct:       r.FeePct,
		SelectedIdx:  bankIdx,
		Total:        r.Total,
	}
}

func computeChart(amount, rate float64, tenure, exit int) loancalc.ChartData {
	exit = clampExit(tenure, exit)
	return loancalc.BuildChart(amount, rate, tenure, exit)
}

func computeStrategyResult(amount, rate float64, tenure, exit, lenderIdx int) loancalc.StrategyResult {
	exit = clampExit(tenure, exit)
	if lenderIdx < 0 || lenderIdx >= len(loancalc.Lenders) {
		lenderIdx = 0
	}
	return loancalc.ComputeStrategy(amount, rate, tenure, exit, true, lenderIdx)
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		stats := computeStats(30000000, 10.5, 60, 18)
		bank := computeBankCost(30000000, 10.5, 60, 18, 0)
		chart := computeChart(30000000, 10.5, 60, 18)
		strategy := computeStrategyResult(30000000, 10.5, 60, 18, 0)
		templ.Handler(templates.Index(stats, bank, strategy, chart)).ServeHTTP(c.Writer, c.Request)
	})

	r.POST("/calculate", func(c *gin.Context) {
		amount, _ := strconv.ParseFloat(c.PostForm("amount"), 64)
		rate, _ := strconv.ParseFloat(c.PostForm("rate"), 64)
		tenure, _ := strconv.Atoi(c.PostForm("tenure"))
		exit, _ := strconv.Atoi(c.PostForm("exit"))
		bankIdx, _ := strconv.Atoi(c.PostForm("bank"))

		stats := computeStats(amount, rate, tenure, exit)
		bank := computeBankCost(amount, rate, tenure, exit, bankIdx)
		chart := computeChart(amount, rate, tenure, exit)
		strategy := computeStrategyResult(amount, rate, tenure, exit, bankIdx)
		templ.Handler(templates.CalcResponse(stats, bank, strategy, chart)).ServeHTTP(c.Writer, c.Request)
	})

	r.Run(":8080")
}
