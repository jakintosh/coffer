package insights

// import (
// 	"bufio"
// 	"encoding/json"
// 	"fmt"
// 	"html/template"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
// 	"time"

// 	"git.sr.ht/~jakintosh/coffer/internal/database"
// )

// type Insights struct {
// 	NumPatrons   int
// 	TotalMonthly int
// }

// func Init(
// 	requests <-chan int,
// ) {

// 	// render page once on startup
// 	renderFundingPage()

// 	go schedulePageRebuilds(requests)
// }

// var insightsJson string = "{}"

// func ServeInsights(w http.ResponseWriter, h *http.Request) {
// 	io.WriteString(w, insightsJson)
// 	w.WriteHeader(http.StatusOK)
// }

// func schedulePageRebuilds(req <-chan int) {
// 	const duration = time.Millisecond * 500
// 	var timer *time.Timer
// 	var c <-chan time.Time
// 	for {
// 		select {
// 		case <-req:
// 			if timer != nil {
// 				timer.Reset(duration)
// 			} else {
// 				timer = time.NewTimer(duration)
// 				c = timer.C
// 			}

// 		case <-c:
// 			c = nil
// 			timer = nil

// 			insights, err := generateInsights()
// 			if err != nil {
// 				// dang
// 			}
// 			bytes, err := json.Marshal(insights)
// 			if err != nil {
// 				// dang
// 			}
// 			insightsJson = string(bytes)
// 		}
// 	}
// }

// func renderFundingPage() {

// 	// generate insights from database
// 	insights, err := generateInsights()
// 	if err != nil {
// 		log.Printf("failed to render funding page: couldn't generate insights: %v\n", err)
// 		return
// 	}

// 	// open writer for output page
// 	f, err := os.Create(FUNDING_PAGE_FILE_PATH)
// 	if err != nil {
// 		log.Printf("failed to render funding page: %v\n", err)
// 		return
// 	}
// 	defer f.Close()
// 	w := bufio.NewWriter(f)

// 	// parse and render template
// 	tmpl, err := template.ParseFiles(FUNDING_PAGE_TMPL_PATH)
// 	if err != nil {
// 		log.Printf("failed to render funding page: %v\n", err)
// 		return
// 	}
// 	err = tmpl.Execute(w, insights)
// 	if err != nil {
// 		log.Printf("failed to render funding page: %v\n", err)
// 		return
// 	}
// 	w.Flush()

// 	log.Println("insights: re-rendered funding page")
// }

// func generateInsights() (*Insights, error) {

// 	summary, err := database.QuerySubscriptionSummary()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to query subscription summary: %v", err)
// 	}
// 	percentGoal := (float64(summary.Total) / float64(MONTHLY_INCOME_GOAL)) * 100
// 	insights := Insights{
// 		NumPatrons:   summary.Count,
// 		TotalMonthly: summary.Total,
// 		PercentGoal:  fmt.Sprintf("%.1f", percentGoal),
// 		Count5:       0,
// 		Count10:      0,
// 		Count20:      0,
// 		Count50:      0,
// 		Count100:     0,
// 	}

// 	for tier, count := range summary.Tiers {
// 		switch tier {
// 		case 5:
// 			insights.Count5 = count
// 		case 10:
// 			insights.Count10 = count
// 		case 20:
// 			insights.Count20 = count
// 		case 50:
// 			insights.Count50 = count
// 		case 100:
// 			insights.Count100 = count
// 		}
// 	}

// 	return &insights, nil
// }
