package cron

// import (
// 	"buffcloud/tracer/internal/com"
// 	"time"
// )

// func DailyNameCodeFailStat() {

// 	comDailyReporter := com.NewMonitorDailyReporter()
// 	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
// 	statMap, err1 := comDailyReporter.DailyNameCodeStat(yesterday)
// 	if err1 != nil {
// 		return
// 	}

// 	comDailyReporter.DailyNameCodeStatReport(yesterday, statMap)

// }

// func DependencyMonitor() {
// 	comDependencyMonitor := com.NewDependencyMonitor()
// 	comDependencyMonitor.Monitor()
// }

// func CleanTracerSpans() {
// 	com, _ := com.NewSpan(nil, nil)
// 	com.Clean()
// }
