package main

import (
	"strconv"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/mayanksekhar/loanexitiq/go-service/loancalc"
	"github.com/mayanksekhar/loanexitiq/go-service/templates"
)

func computeStats(amount, rate float64, tenure, exit int) templates.StatsData {
	if exit > tenure-3 {
		exit = tenure - 3
	}
	if exit < 1 {
		exit = 1
	}

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

func main() {
	r := gin.Default()
	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		stats := computeStats(30000000, 10.5, 60, 18)
		templ.Handler(templates.Index(stats)).ServeHTTP(c.Writer, c.Request)
	})

	r.POST("/calculate", func(c *gin.Context) {
		amount, _ := strconv.ParseFloat(c.PostForm("amount"), 64)
		rate, _ := strconv.ParseFloat(c.PostForm("rate"), 64)
		tenure, _ := strconv.Atoi(c.PostForm("tenure"))
		exit, _ := strconv.Atoi(c.PostForm("exit"))

		stats := computeStats(amount, rate, tenure, exit)
		templ.Handler(templates.Stats(stats)).ServeHTTP(c.Writer, c.Request)
	})

	r.Run(":8080")
}
