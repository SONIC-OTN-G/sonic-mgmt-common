//////////////////////////////////////////////////////////////////////////
//
// OTN 告警上报(方案一)—— CURALARM 表扫描/聚合
// CuralarmStore 抽象键值读取:生产实现在 translib/alarm_app.go 胶水层
// (包装 translib/db);本包测试用 fake,免 redis。
//
//////////////////////////////////////////////////////////////////////////

package alarm

import (
	"sort"
	"strconv"
)

const curalarmPrefix = "CURALARM"

// CuralarmStore 抽象 CURALARM 表读取。
// Keys() 返回表内全部裸键名(resource#type-id,不含表前缀)。
type CuralarmStore interface {
	Keys() ([]string, error)
	HGetAll(key string) (map[string]string, error)
}

// GetAlarmEntries 扫描 CURALARM 全表 → 逐条翻译 → 按 time-created 降序。
// 单条翻译失败(如缺 time-created)跳过该条,不拖垮整表;
// store 层错误原样上抛。
func GetAlarmEntries(st CuralarmStore) ([]map[string]interface{}, error) {
	keys, err := st.Keys()
	if err != nil {
		return nil, err
	}

	type item struct {
		ts int64
		e  map[string]interface{}
	}
	items := make([]item, 0, len(keys))

	for _, key := range keys {
		fields, err := st.HGetAll(key)
		if err != nil {
			return nil, err
		}
		e, err := TranslateCuralarmEntry(fields, key)
		if err != nil {
			continue // 坏条目跳过
		}
		ts, _ := strconv.ParseInt(fields["time-created"], 10, 64)
		items = append(items, item{ts: ts, e: e})
	}

	sort.Slice(items, func(i, j int) bool { return items[i].ts > items[j].ts })

	entries := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		entries = append(entries, it.e)
	}
	return entries, nil
}

// Prefix 供胶水层拼表名使用
func Prefix() string { return curalarmPrefix }
