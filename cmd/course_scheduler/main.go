// main.go
package main

import (
	"course_scheduler/internal/base"
	"course_scheduler/internal/genetic_algorithm"
	"course_scheduler/internal/utils"
	"log"
	"time"
)

func main() {
	// 创建日志文件
	logFile := utils.SetUpLogFile()
	defer logFile.Close()

	// 开始时间
	startTime := time.Now()

	// 监控器
	monitor := base.NewMonitor()

	// 加载测试数据
	configFilePath := "/Users/apple/Documents/work/my/course_scheduler/testdata/grade_school.yaml"
	scheduleInput, err := base.LoadTestData(configFilePath)
	if err != nil {
		log.Fatalf("load test data failed. %s", err)
	}

	// 检查输入数据
	if err := scheduleInput.Check(); err != nil {

		log.Fatalf("check teach task allocation failed. %s", err)
	}

	// 遗传算法排课
	bestIndividual, bestGen, err := genetic_algorithm.Execute(scheduleInput, monitor, startTime)
	if err != nil {
		log.Fatalf("genetic execute failed. %s", err)
	}

	// 结束时间
	monitor.TotalTime = time.Since(startTime)

	// 输出最终排课结果
	log.Println("🍻 Best solution done!")

	// 打印最好的个体
	log.Printf("bestGen: %d, bestIndividual.Fitness: %d, uniqueId: %s\n", bestGen, bestIndividual.Fitness, bestIndividual.UniqueId)
	bestIndividual.PrintSchedule(scheduleInput.Schedule, scheduleInput.Subjects)

	// 打印个体的约束状态信息
	log.Println("打印个体的约束状态信息")
	bestIndividual.PrintConstraints()

	// 打印监控数据
	monitor.Dump()
}
