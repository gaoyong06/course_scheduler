/*
 * @Author: gaoyong gaoyong79@gogoep.com
 * @Date: 2024-07-11 12:00:57
 * @LastEditors: gaoyong gaoyong79@gogoep.com
 * @LastEditTime: 2024-07-11 13:30:12
 * @FilePath: \course_scheduler\internal\models\class.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
// class.go
package models

type Class struct {
	SchoolID int    `json:"school_id" mapstructure:"school_id"`
	ClassID  int    `json:"class_id" mapstructure:"class_id"`
	Name     string `json:"name" mapstructure:"name"`
}
