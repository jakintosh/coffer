package main

import (
	"bufio"
	"errors"
	"fmt"
	"html/template"
	"log"
	"os"
)

type Insights struct {
	NumPatrons   int
	TotalMonthly int
	PercentGoal  string
	Count5       int
	Count10      int
	Count20      int
	Count50      int
	Count100     int
}

func renderFundingPage() {

	// generate insights from database
	insights, err := generateInsights()
	if err != nil {
		log.Println("insights: failed to render funding page: couldn't generate insights")
		return
	}

	// open writer for output page
	f, err := os.Create(FUNDING_PAGE_FILE_PATH)
	if err != nil {
		log.Printf("insights: failed to render funding page: %v\n", err)
		return
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	// parse and render template
	tmpl, err := template.ParseFiles(FUNDING_PAGE_TMPL_PATH)
	if err != nil {
		log.Printf("insights: failed to render funding page: %v\n", err)
		return
	}
	err = tmpl.Execute(w, insights)
	if err != nil {
		log.Printf("insights: failed to render funding page: %v\n", err)
		return
	}
	w.Flush()

	log.Println("insights: re-rendered funding page")
}

func generateInsights() (Insights, error) {

	numPatrons, totalMonthly, tierCounts, err := scanSubscriptions()
	if err != nil {
		log.Println("failed to generate insights")
		return Insights{}, err
	}
	percentGoal := (float64(totalMonthly) / float64(MONTHLY_INCOME_GOAL)) * 100
	insights := Insights{
		NumPatrons:   numPatrons,
		TotalMonthly: totalMonthly,
		PercentGoal:  fmt.Sprintf("%.1f", percentGoal),
		Count5:       0,
		Count10:      0,
		Count20:      0,
		Count50:      0,
		Count100:     0,
	}

	for tier, count := range tierCounts {
		switch tier {
		case 5:
			insights.Count5 = count
		case 10:
			insights.Count10 = count
		case 20:
			insights.Count20 = count
		case 50:
			insights.Count50 = count
		case 100:
			insights.Count100 = count
		}
	}

	return insights, nil
}

func scanSubscriptions() (int, int, map[int]int, error) {

	var count int
	var amount int
	tierCounts := make(map[int]int)

	// query summary info
	summary_statement := `
		SELECT COUNT(*), COALESCE(SUM(amount), 0)
		FROM subscription
		WHERE status='active'
		AND currency='usd';`
	rows, err := db.Query(summary_statement)
	if err != nil {
		log.Printf("insights: failed to query summary_statement: %v\n", err)
		return -1, -1, nil, err
	}

	// scan summary rows
	defer rows.Close()
	if !rows.Next() {
		return -1, -1, nil, errors.New("sql failure: subscription summary query")
	}
	err = rows.Scan(&count, &amount)
	if err != nil {
		return -1, -1, nil, errors.New(fmt.Sprintf("failed to scan row of summary_statement: %v\n", err))
	}

	// query tier info
	tier_statement := `
		SELECT amount, COUNT(*)
		FROM subscription
		WHERE status='active'
		AND currency='usd'
		GROUP BY amount;`
	rows, err = db.Query(tier_statement)
	if err != nil {
		log.Printf("insights: failed to query tier_statement: %s\n", err)
		return -1, -1, nil, err
	}

	// scan tier rows
	for rows.Next() {
		var amount int
		var count int
		err := rows.Scan(&amount, &count)
		if err != nil {
			log.Printf("insights: failed to scan row of tier_statement: %s\n", err)
			return -1, -1, nil, err
		}
		tierCounts[(amount / 100)] = count
	}

	return count, amount / 100, tierCounts, nil
}

func scanPayments() map[string]int {
	statement := `
		SELECT COALESCE(SUM(amount), 0) as amount,
			strftime('%m-%Y', DATETIME(created, 'unixepoch')) AS 'month-year'
		FROM payment
		WHERE currency='usd'
		AND status='succeeded'
		GROUP BY 'month-year';`

	rows, err := db.Query(statement)
	if err != nil {
		// TODO: error handling
	}

	monthlyPayments := make(map[string]int)
	defer rows.Close()
	for rows.Next() {
		var amount int
		var date string
		err := rows.Scan(&amount, &date)
		if err != nil {
			// TODO: error handling
		}
		monthlyPayments[date] = amount / 100
	}

	return monthlyPayments
}
