//////////////////////////////////////////////////////////////////////////
//
// OTN 告警上报(方案一)—— GetAlarmEntries 扫描/聚合单测(fake store,免 redis)
//
//////////////////////////////////////////////////////////////////////////

package alarm

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type fakeStore struct {
	keys    []string
	entries map[string]map[string]string
	keyErr  error
}

func (f *fakeStore) Keys() ([]string, error) {
	if f.keyErr != nil {
		return nil, f.keyErr
	}
	return f.keys, nil
}

func (f *fakeStore) HGetAll(key string) (map[string]string, error) {
	return f.entries[key], nil
}

func entry(resource, typeID, severity, text string, ts int64) map[string]interface{} {
	return map[string]interface{}{
		"resource":      resource,
		"alarm-type-id": typeID,
		"severity":      severity,
		"text":          text,
		"time-created":  time.Unix(0, ts).UTC().Format(time.RFC3339),
		"status":        "RAISED",
	}
}

func TestGetAlarmEntries(t *testing.T) {
	fixed := int64(1792843160178746605)
	earlier := fixed - 1000000000 // 早 1 秒

	st := &fakeStore{
		keys: []string{"linecard1#LOW_TEMPERATURE_ALARM", "linecard1#HIGH_TEMPERATURE_ALARM"},
		entries: map[string]map[string]string{
			"linecard1#LOW_TEMPERATURE_ALARM": {
				"id": "linecard1#LOW_TEMPERATURE_ALARM", "time-created": itoa(fixed),
				"text": "low temp", "severity": "WARNING",
				"type-id": "LOW_TEMPERATURE_ALARM", "resource": "linecard1",
			},
			"linecard1#HIGH_TEMPERATURE_ALARM": {
				"id": "linecard1#HIGH_TEMPERATURE_ALARM", "time-created": itoa(earlier),
				"text": "high temp", "severity": "MAJOR",
				"type-id": "HIGH_TEMPERATURE_ALARM", "resource": "linecard1",
			},
		},
	}

	got, err := GetAlarmEntries(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 entries, got %d: %v", len(got), got)
	}
	// 降序:新的(LOW)在前
	if got[0]["text"] != "low temp" || got[1]["text"] != "high temp" {
		t.Fatalf("sort desc broken: %v", got)
	}
	if !reflect.DeepEqual(got[0], entry("linecard1", "LOW_TEMPERATURE_ALARM", "WARNING", "low temp", fixed)) {
		t.Fatalf("entry0 mismatch: %v", got[0])
	}
}

func TestGetAlarmEntriesEmpty(t *testing.T) {
	st := &fakeStore{keys: []string{}, entries: map[string]map[string]string{}}
	got, err := GetAlarmEntries(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("want non-nil empty slice, got %#v", got)
	}
}

// 坏条目(缺 time-created)跳过不拖垮整表
func TestGetAlarmEntriesSkipsInvalidEntry(t *testing.T) {
	fixed := int64(1792843160178746605)
	st := &fakeStore{
		keys: []string{"linecard1#OK", "#BAD"},
		entries: map[string]map[string]string{
			"linecard1#OK": {
				"id": "linecard1#OK", "time-created": itoa(fixed),
				"text": "ok", "severity": "MINOR",
				"type-id": "OK", "resource": "linecard1",
			},
			"#BAD": {
				// 缺 time-created
				"id": "#BAD", "text": "bad", "severity": "MAJOR",
				"type-id": "BAD", "resource": "",
			},
		},
	}
	got, err := GetAlarmEntries(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0]["text"] != "ok" {
		t.Fatalf("want only the valid entry, got %v", got)
	}
}

func TestGetAlarmEntriesStoreError(t *testing.T) {
	st := &fakeStore{keyErr: errFake}
	if _, err := GetAlarmEntries(st); err != errFake {
		t.Fatalf("want store error %v, got %v", errFake, err)
	}
}

var errFake = errorString("fake store error")

func itoa(i int64) string { return fmt.Sprintf("%d", i) }

type errorString string

func (e errorString) Error() string { return string(e) }
