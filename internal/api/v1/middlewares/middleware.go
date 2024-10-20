package middlewares

import (
	"course_scheduler/internal/base"
	"course_scheduler/internal/genetic_algorithm"
	"course_scheduler/internal/models"
	"course_scheduler/internal/types"
	"course_scheduler/internal/utils"
	"fmt"
	"log"
	"time"
)

// 执行排课任务
// 排课的逻辑写在中间件中原因如下
// 1. 分离关注点：中间件是处理请求和响应之间的中间逻辑的组件，将排课的逻辑写在中间件中可以将业务逻辑与 HTTP 处理程序分离开来，使得代码更加模块化、易于维护和扩展
// 2. 重用性：中间件可以在多个处理程序中重用，如果将排课的逻辑写在中间件中，那么可以在不同的处理程序中重用该逻辑，提高代码的重用性
// 3. 可测试性：将排课的逻辑写在中间件中可以提高代码的可测试性，因为中间件可以独立于 HTTP 处理程序进行测试，这使得测试排课的逻辑更加方便和高效
func ExecuteTask(taskID uint64, taskData string) ([]*models.ScheduleResult, int, error) {
	// 创建日志文件
	logFile := utils.SetUpLogFile()
	defer logFile.Close()

	// 开始时间
	startTime := time.Now()

	// 监控器
	monitor := base.NewMonitor()

	// 加载测试数据
	scheduleInput, err := base.ParseScheduleInputFromJSON(taskData)
	if err != nil {
		return nil, 0, fmt.Errorf("load test data failed. %s", err)
	}

	// 检查输入数据
	err = scheduleInput.Check()
	if err != nil {
		return nil, 0, fmt.Errorf("check teach task allocation failed. %s", err)
	}

	// 遗传算法排课
	bestIndividual, bestGen, err := genetic_algorithm.Execute(scheduleInput, monitor, startTime)
	if err != nil {
		return nil, 0, fmt.Errorf("genetic execute failed. %s", err)
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
	// monitor.Dump()

	// 将 bestIndividual 转换为 []*models.ScheduleResult
	scheduleResults, err := convertIndividualToScheduleResults(taskID, bestIndividual, scheduleInput)
	if err != nil {
		return nil, 0, err
	}

	return scheduleResults, bestGen, nil
}

// 将遗传个体类型转换为排课结果类型
func convertIndividualToScheduleResults(taskID uint64, individual *genetic_algorithm.Individual, input *base.ScheduleInput) ([]*models.ScheduleResult, error) {

	var scheduleResults []*models.ScheduleResult

	// 根据 individual 和 scheduleInput 的数据结构，进行数据的转换和处理
	for _, chromosomes := range individual.Chromosomes {

		for _, gene := range chromosomes.Genes {

			SN, err := types.ParseSN(gene.ClassSN)
			if err != nil {
				return nil, err
			}

			totalClassesPerDay := input.Schedule.GetTotalClassesPerDay()

			for _, timeSlot := range gene.TimeSlots {

				weekday := int8(timeSlot / totalClassesPerDay)
				period := int8(timeSlot % totalClassesPerDay)

				result := &models.ScheduleResult{
					TaskID:    taskID,
					SubjectID: uint64(SN.SubjectID),
					TeacherID: uint64(gene.TeacherID),
					GradeID:   uint64(SN.GradeID),
					ClassID:   uint64(SN.ClassID),
					VenueID:   uint64(gene.VenueID),
					Weekday:   weekday,
					Period:    period,
					// TODO: 待完成
					// StartTime: item.StartTime,
					// EndTime:   item.EndTime,
				}
				scheduleResults = append(scheduleResults, result)

			}
		}
	}

	return scheduleResults, nil
}
