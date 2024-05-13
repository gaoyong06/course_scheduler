// selection.go
package genetic_algorithm

import (
	"course_scheduler/config"
	"log"
	"time"
)

// 根据终止条件判断是否终止进化搜索循环
// 比如达到最大迭代次数或找到满意的解等
// 返回 true 表示终止搜索循环，返回 false 表示继续搜索循环
// ...
// currentIteration 当前迭代次数
// foundSatSolution 找到满意的解
// genWithoutImprovement 连续 n 代没有改进
// startTime 当前时间
func TerminationCondition(currentIteration int, foundSatSolution bool, genWithoutImprovement int, startTime time.Time) bool {
	// 达到最大迭代次数
	if currentIteration >= config.MaxGen {
		log.Println("Termination condition: Reached maximum iteration.")
		return true
	}

	// 找到满意的解
	if foundSatSolution {
		log.Println("Termination condition: Found satisfactory solution.")
		return true
	}

	// 连续 n 代没有改进
	if genWithoutImprovement >= config.MaxStagnGen {
		log.Println("Termination condition: Reached maximum generations without improvement.")
		return true
	}

	// 达到预先定义的总运行时间
	if time.Since(startTime) >= config.MaxDuration {
		log.Println("Termination condition: Reached maximum running duration.")
		return true
	}

	return false
}
