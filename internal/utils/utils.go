package utils

import (
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

type ClassMatrixItem struct {
	ClassSN  string
	TimeSlot int
	Day      int
	Period   int
	Score    int
}

// PrintClassMatrix 以Markdown格式打印classMatrix
func PrintClassMatrix(classMatrix map[string]map[int]map[int]map[int]types.Val, schedule *models.Schedule) {

	fmt.Println("| Time Slot| Day | Period | Score |")
	fmt.Println("| --- | --- | --- | --- |")

	items := make([]ClassMatrixItem, 0)
	totalClassesPerDay := schedule.GetTotalClassesPerDay()

	for classSN, teacherMap := range classMatrix {
		if classSN == "1_1_1" {
			for _, venueMap := range teacherMap {
				for _, timeSlotMap := range venueMap {
					for timeSlot, val := range timeSlotMap {
						day := timeSlot / totalClassesPerDay
						period := timeSlot % totalClassesPerDay
						item := ClassMatrixItem{
							ClassSN:  classSN,
							TimeSlot: timeSlot,
							Day:      day,
							Period:   period,
							Score:    val.ScoreInfo.Score,
						}
						items = append(items, item)
					}
				}
			}
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Day == items[j].Day {
			return items[i].Period < items[j].Period
		}
		return items[i].Day < items[j].Day
	})

	for _, item := range items {
		fmt.Printf("| %d | %d | %d | %d |\n", item.TimeSlot, item.Day, item.Period, item.Score)
	}
}

// 设置日志文件
// 日志文件保存在logs目录下,并按年月日分开保存
func SetUpLogFile() *os.File {
	// 创建 logs 目录
	logDir := "../logs"
	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		panic(err)
	}

	// 获取当前日期和时间
	now := time.Now()
	year := now.Year()
	month := now.Month()
	day := now.Day()

	// 构建日志文件路径
	logFilename := "log_" + strconv.Itoa(year) + strconv.Itoa(int(month)) + strconv.Itoa(day) + ".txt"
	logFilePath := filepath.Join(logDir, logFilename)

	// 创建日志文件
	logFile, err := os.Create(logFilePath)
	if err != nil {
		panic(err)
	}

	// 设置日志输出到文件
	// log.SetOutput(logFile)
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	return logFile
}
