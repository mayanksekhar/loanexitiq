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

func computeLeaderboard(amount, rate float64, tenure, exit int) []templates.LenderResult {
	exit = clampExit(tenure, exit)
	results := loancalc.ComputeLeaderboard(amount, rate, tenure, exit, true)
	out := make([]templates.LenderResult, len(results))
	for i, r := range results {
		out[i] = templates.LenderResult{
			Name:         r.Name,
			Note:         r.Note,
			RatePct:      r.RatePct,
			InterestPaid: r.InterestPaid,
			Fee:          r.Fee,
			Total:        r.Total,
			IsBest:       r.IsBest,
			IsWorst:      r.IsWorst,
		}
	}
	return out
}

func computeStrategyResult(amount, rate float64, tenure, exit, lenderIdx int) loancalc.StrategyResult {
	exit = clampExit(tenure, exit)
	if lenderIdx < 0 || lenderIdx >= len(loancalc.StrategyLenders) {
		lenderIdx = 0
	}
	return loancalc.ComputeStrategy(amount, rate, tenure, exit, true, lenderIdx)
}

func main() {
	r := gin.Default()
	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		stats := computeStats(30000000, 10.5, 60, 18)
		results := computeLeaderboard(30000000, 10.5, 60, 18)
		strategy := computeStrategyResult(30000000, 10.5, 60, 18, 0)
		templ.Handler(templates.Index(stats, results, strategy)).ServeHTTP(c.Writer, c.Request)
	})

	r.POST("/calculate", func(c *gin.Context) {
		amount, _ := strconv.ParseFloat(c.PostForm("amount"), 64)
		rate, _ := strconv.ParseFloat(c.PostForm("rate"), 64)
		tenure, _ := strconv.Atoi(c.PostForm("tenure"))
		exit, _ := strconv.Atoi(c.PostForm("exit"))
		lenderIdx, _ := strconv.Atoi(c.PostForm("lender"))

		stats := computeStats(amount, rate, tenure, exit)
		results := computeLeaderboard(amount, rate, tenure, exit)
		strategy := computeStrategyResult(amount, rate, tenure, exit, lenderIdx)
		templ.Handler(templates.CalcResponse(stats, results, strategy)).ServeHTTP(c.Writer, c.Request)
	})

	r.Run(":8080")
}
