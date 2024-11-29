package insights

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"os"
	"time"

	"git.sr.ht/~jakintosh/studiopollinator-api/internal/database"
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

var FUNDING_PAGE_FILE_PATH string
var FUNDING_PAGE_TMPL_PATH string
var MONTHLY_INCOME_GOAL int

func Init(
	filepath string,
	templatePath string,
	monthlyGoal int,
	requests <-chan int,
) {

	FUNDING_PAGE_FILE_PATH = filepath
	FUNDING_PAGE_TMPL_PATH = templatePath
	MONTHLY_INCOME_GOAL = monthlyGoal

	go schedulePageRebuilds(requests)
}

func schedulePageRebuilds(req <-chan int) {
	var timer *time.Timer = nil
	var c <-chan time.Time = nil
	duration := time.Millisecond * 500
	for {
		select {
		case <-req:
			if timer != nil {
				timer.Reset(duration)
			} else {
				timer = time.NewTimer(duration)
				c = timer.C
			}

		case <-c:
			c = nil
			timer = nil
			renderFundingPage()
		}
	}
}

func renderFundingPage() {

	// generate insights from database
	insights, err := generateInsights()
	if err != nil {
		log.Printf("insights: failed to render funding page: couldn't generate insights: %v\n", err)
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

func generateInsights() (*Insights, error) {

	summary, err := database.QuerySubscriptionSummary()
	if err != nil {
		return nil, err
	}
	percentGoal := (float64(summary.Total) / float64(MONTHLY_INCOME_GOAL)) * 100
	insights := Insights{
		NumPatrons:   summary.Count,
		TotalMonthly: summary.Total,
		PercentGoal:  fmt.Sprintf("%.1f", percentGoal),
		Count5:       0,
		Count10:      0,
		Count20:      0,
		Count50:      0,
		Count100:     0,
	}

	for tier, count := range summary.Tiers {
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

	return &insights, nil
}
