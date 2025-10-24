package service

import (
	"regexp"
	"strings"
	"time"
)

// MIKU-1660 [lived] 修复streamconf whep相关字段json tag (#2458)
func ParseTitle(mrTitle string) (GiraId string, changeModules string, err error) {
	GiraId = ParseGiraIdFromTitle(mrTitle)

	// 定义正则表达式，匹配中括号内的内容
	re := regexp.MustCompile(`\[(.*?)\]`)

	// 查找所有匹配的内容
	matches := re.FindAllStringSubmatch(mrTitle, -1)
	if len(matches) > 0 {
		changeModules = matches[0][1]
	}
	return GiraId, changeModules, nil
}

func ParseGiraIdFromTitle(mrTitle string) string {
	if !strings.HasPrefix(mrTitle, "MIKU-") {
		return ""
	}
	parts := strings.Split(mrTitle, " ")
	if len(parts) < 2 {
		return ""
	}
	matchkey := parts[0][5:]
	return matchkey
}

func TimeToBeijing(t time.Time) string {
	// 将时间转换为北京时间（UTC+8）
	beijingTime := t.In(time.FixedZone("CST", 8*3600))
	// 返回格式化后的时间字符串
	return beijingTime.Format("2006-01-02 15:04:05")
}
