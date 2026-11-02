//////////////////////////////////////////////////////////////////////////
//
// OTN 告警上报(方案一)—— CURALARM → openconfig-alarms 翻译纯函数
// 无外部依赖;供 translib/alarm_app(GET)与 netconfd notification 复用
//
//////////////////////////////////////////////////////////////////////////

package alarm

import (
	"fmt"
	"strconv"
	"time"
)

// severity 直通映射;未知值一律 UNKNOWN,不 panic(Review Focus 2)
var severityMap = map[string]string{
	"MINOR":    "MINOR",
	"WARNING":  "WARNING",
	"MAJOR":    "MAJOR",
	"CRITICAL": "CRITICAL",
}

// TranslateCuralarmEntry 把一条 CURALARM hash 字段翻译成 openconfig-alarms
// alarm 节点(fields 为六字段,key 为键名 resource#type-id)。
// time-created 缺失/非法即报错——没有产生时间的告警不可上报。
func TranslateCuralarmEntry(fields map[string]string, key string) (map[string]interface{}, error) {
	ns, ok := fields["time-created"]
	if !ok {
		return nil, fmt.Errorf("curalarm %s: missing time-created", key)
	}
	ts, err := strconv.ParseInt(ns, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("curalarm %s: bad time-created %q: %w", key, ns, err)
	}

	severity := fields["severity"]
	if mapped, ok := severityMap[severity]; ok {
		severity = mapped
	} else {
		severity = "UNKNOWN"
	}

	return map[string]interface{}{
		"resource":      fields["resource"],
		"alarm-type-id": fields["type-id"],
		"severity":      severity,
		"text":          fields["text"],
		"time-created":  time.Unix(0, ts).UTC().Format(time.RFC3339),
		"status":        "RAISED",
	}, nil
}
