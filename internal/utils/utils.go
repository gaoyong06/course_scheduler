package utils

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

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
