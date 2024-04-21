// score_cache.go
package evaluation

// import (
// 	"strconv"
// 	"sync"
// )

// // 缓存计算结果，key为sn+teacherID+venueID+timeSlot，value为计算结果
// var calcScoreCache = make(map[string]CalcScoreResult)
// var calcScoreCacheLock sync.RWMutex

// // 获取缓存计算结果
// func getCachedScore(sn string, teacherID, venueID, timeSlot int) (*CalcScoreResult, bool) {
// 	key := sn + "_" + strconv.Itoa(teacherID) + "_" + strconv.Itoa(venueID) + "_" + strconv.Itoa(timeSlot)
// 	calcScoreCacheLock.RLock()
// 	score, ok := calcScoreCache[key]
// 	calcScoreCacheLock.RUnlock()
// 	return &score, ok
// }

// // 保存缓存计算结果
// func saveCachedScore(sn string, teacherID, venueID, timeSlot int, result CalcScoreResult) {
// 	key := sn + "_" + strconv.Itoa(teacherID) + "_" + strconv.Itoa(venueID) + "_" + strconv.Itoa(timeSlot)
// 	calcScoreCacheLock.Lock()
// 	defer calcScoreCacheLock.Unlock() // 在函数退出时释放锁
// 	calcScoreCache[key] = result
// }
