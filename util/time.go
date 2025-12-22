package util

import (
	"time"
)

func Millisecond() uint32 {
	ms := time.Now().UnixNano() / 1e6
	return uint32(ms)
}
func Second() int64 {
	return time.Now().Unix()
}

func Second2Base62() string {
	return Number2String(Second(), Alphanumeric)
}

var TimestampFormat = "2006-01-02 15:04:05"

func FormatTime(tm time.Time) string {
	return tm.Format(TimestampFormat)
}

var DateFormat = "2006-01-02"

func FormatDate(tm time.Time) string {
	return tm.Format(DateFormat)
}

func ParseFormatTime(timeStr, timestampFormat string) (time.Time, error) {
	parse, err := time.Parse(timestampFormat, timeStr)
	if err != nil {
		return parse, err
	}
	return parse, nil
}

func NowDateTime() string {
	return FormatTime(time.Now())
}
func NowDateFormatTime(timestampFormat string) string {
	return time.Now().Format(timestampFormat)
}
func GetNowTime() time.Time {

	return time.Now()

}

func IsAfter(pre string, now time.Time, timestampFormat string) bool {
	formatTime, err := ParseFormatTime(pre, timestampFormat)
	if err != nil {
		return false
	}
	return now.After(formatTime)

}
