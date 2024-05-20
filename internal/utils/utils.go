package utils

import (
	"course_scheduler/internal/models"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cast"
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

// 将字符串a_b,转为[a,b]
func ParseTimeSlotStr(timeSlotStr string) []int {

	parts := strings.Split(timeSlotStr, "_")
	timeSlots := make([]int, len(parts))
	for i, part := range parts {
		timeSlot := cast.ToInt(part)
		timeSlots[i] = timeSlot
	}
	return timeSlots
}

// 将数组[a,b]转为字符串a_b
func TimeSlotsToStr(timeSlot []int) string {

	// 将时间段切片中的元素转换为字符串
	strs := make([]string, len(timeSlot))
	for i, t := range timeSlot {
		strs[i] = strconv.Itoa(t)
	}

	// 连接字符串并返回
	return strings.Join(strs, "_")
}

// splitSlice 将一个整数切片按照指定的组数进行分组
func SplitSlice(slice []int, groupSize int) [][]int {
	// 计算出分组后的切片的长度
	groups := len(slice) / groupSize
	if len(slice)%groupSize != 0 {
		groups++
	}

	// 创建一个用于存储分组后的切片的切片
	result := make([][]int, 0, groups)

	// 遍历原始切片，按照指定的组数进行分组
	for i := 0; i < len(slice); i += groupSize {
		end := i + groupSize
		if end > len(slice) {
			end = len(slice)
		}
		group := slice[i:end]
		result = append(result, group)
	}

	return result
}

// GroupByPairs 将一个整数切片按照指定的对数进行分组
// slice:[0,1,2,3,4,5],pairs:1 result: [[0,1],[2],[3],[4],[5]]
func GroupByPairs(slice []int, pairs int) ([][]int, error) {

	// 计算出分组后的切片的长度
	if pairs*2 > len(slice) {
		return nil, fmt.Errorf("the length of slice is less than %d", pairs*2)
	}

	// 创建一个用于存储分组后的切片的切片
	result := make([][]int, 0, len(slice))
	end := pairs * 2

	for i := 0; i < len(slice); i++ {

		if i < end {
			group := make([]int, 0, 2)
			group = append(group, slice[i])
			group = append(group, slice[i+1])
			result = append(result, group)
			i++
		} else {
			group := make([]int, 0, 1)
			group = append(group, slice[i])
			result = append(result, group)
		}
	}
	return result, nil
}

// GroupedIntsToString 将一个分组的整数切片转换为字符串切片
// slice: [[0,1],[2],[3],[4],[5]], result: ["0_1", "2", "3", "4","5"]
func GroupedIntsToString(slice [][]int) ([]string, error) {
	// 创建一个用于存储转换后的字符串切片的切片
	result := make([]string, 0, len(slice))

	// 遍历分组的整数切片，将每个分组转换为字符串
	for _, group := range slice {
		// 如果分组中只有一个元素，则将其转换为字符串
		if len(group) == 1 {
			str := strconv.Itoa(group[0])
			result = append(result, str)
		} else if len(group) == 2 {
			// 如果分组中有两个元素，则将其连接成一个字符串
			str := strconv.Itoa(group[0]) + "_" + strconv.Itoa(group[1])
			result = append(result, str)
		} else {
			// 如果分组中有多于两个元素，则返回一个错误
			return nil, fmt.Errorf("the length of group is more than 2")
		}
	}

	return result, nil
}

// 从availableSlots获取一个可用的连堂课时间
func GetConnectedTimeSlots(schedule *models.Schedule, availableSlots []int) (int, int) {

	totalClassesPerDay := schedule.GetTotalClassesPerDay()

	// 找出所有的连堂课时间段
	pairs := make([][]int, 0)
	for i := 0; i < len(availableSlots)-1; i++ {
		if availableSlots[i]+1 == availableSlots[i+1] {
			pairs = append(pairs, []int{availableSlots[i], availableSlots[i+1]})
		}
	}

	// 上午的时间段
	forenoonStartPeriod, forenoonEndPeriod := schedule.GetPeriodWithRange("forenoon")

	// 下午的时间段
	afternoonStartPeriod, afternoonEndPeriod := schedule.GetPeriodWithRange("afternoon")

	// 遍历所有的连堂课时间段，找出一个可用的
	for _, pair := range pairs {

		period0 := pair[0] % totalClassesPerDay
		period1 := pair[1] % totalClassesPerDay

		if (period0 >= forenoonStartPeriod && period1 <= forenoonEndPeriod) ||
			(period0 >= afternoonStartPeriod && period1 <= afternoonEndPeriod) {
			return pair[0], pair[1]
		}
	}

	// 没有找到可用的连堂课时间，返回-1，-1
	return -1, -1
}
