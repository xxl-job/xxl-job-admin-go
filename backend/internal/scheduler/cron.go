// Package scheduler 调度核心：cron 解析、时间轮、快慢线程池、调度循环。
package scheduler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"
)

// 调度类型
const (
	ScheduleNone     int8 = 0
	ScheduleCron     int8 = 1 // CRON 表达式（6 字段含秒）
	ScheduleFixRate  int8 = 2 // 固定频率（秒）
	ScheduleFixDelay int8 = 3 // 固定延时（秒，P2 简化为从触发时算）
)

// cronParser 6 字段含秒的 cron parser（与 xxl-job 风格一致）。
var cronParser = cron.NewParser(
	cron.Second | cron.Minute | cron.Hour |
		cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

// NextTriggerTime 根据调度类型计算下次触发时间（毫秒），from 为基准时间。
// 固定延时严格应从上次执行完成算，P2 简化为从基准时间算，P5 完善。
func NextTriggerTime(scheduleType int8, conf string, from time.Time) (int64, error) {
	switch scheduleType {
	case ScheduleCron:
		sched, err := cronParser.Parse(conf)
		if err != nil {
			return 0, fmt.Errorf("parse cron %q: %w", conf, err)
		}
		return sched.Next(from).UnixMilli(), nil
	case ScheduleFixRate, ScheduleFixDelay:
		sec, err := strconv.Atoi(conf)
		if err != nil || sec <= 0 {
			return 0, fmt.Errorf("invalid fixed interval %q", conf)
		}
		return from.Add(time.Duration(sec) * time.Second).UnixMilli(), nil
	default:
		return 0, fmt.Errorf("unsupported schedule type %d", scheduleType)
	}
}
