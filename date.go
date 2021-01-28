// Package pocket Create at 2020-11-06 10:17
package pocket

import (
	"log"
	"strconv"
	"time"
)

// UnixSecond unix time second
func UnixSecond() int64 {
	return time.Now().Unix()
}

// UnixMillisecond unix time millisecond
func UnixMillisecond() int64 {
	return time.Now().UnixNano() / 1e6
}

// GetYear year of given time
func GetYear(t time.Time) int {
	return t.Year()
}

// GetMonth month of given time
func GetMonth(t time.Time) int {
	month, err := strconv.Atoi(t.Format("01"))
	if nil != err {
		log.Println(err.Error())
		return -1
	}
	return month
}

// GetDay day of given time
func GetDay(t time.Time) int {
	return t.Day()
}

// GetDaysOfMonth days in month
func GetDaysOfMonth(year int, month int) (days int) {
	if month != 2 {
		if month == 4 || month == 6 || month == 9 || month == 11 {
			days = 30
		} else {
			days = 31
		}
	} else {
		if ((year%4) == 0 && (year%100) != 0) || (year%400) == 0 {
			days = 29
		} else {
			days = 28
		}
	}
	return
}

// GetDaysOfMonthByTime days in month
func GetDaysOfMonthByTime(t time.Time) (days int) {
	year := t.Year()
	month := GetMonth(t)
	if month != 2 {
		if month == 4 || month == 6 || month == 9 || month == 11 {
			days = 30
		} else {
			days = 31
		}
	} else {
		if ((year%4) == 0 && (year%100) != 0) || (year%400) == 0 {
			days = 29
		} else {
			days = 28
		}
	}
	return
}
