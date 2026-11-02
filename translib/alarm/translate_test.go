//////////////////////////////////////////////////////////////////////////
//
// OTN 告警上报(方案一)—— alarm 翻译纯函数单测
// 覆盖 Review Focus:CURALARM 键特殊字符(#)、severity 越界、ns epoch 时间戳
//
//////////////////////////////////////////////////////////////////////////

package alarm

import (
	"reflect"
	"testing"
	"time"
)

func TestTranslateCuralarmEntry(t *testing.T) {
	// 固定 ns epoch:2026-10-24T11:59:20.178746605Z(UTC)
	fixedNS := int64(1792843160178746605)
	wantTime := time.Unix(0, fixedNS).UTC().Format(time.RFC3339)

	cases := []struct {
		name    string
		key     string
		fields  map[string]string
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "normal entry",
			key:  "linecard1#LOW_TEMPERATURE_ALARM",
			fields: map[string]string{
				"id":           "linecard1#LOW_TEMPERATURE_ALARM",
				"time-created": "1792843160178746605",
				"text":         "injected low temperature",
				"severity":     "WARNING",
				"type-id":      "LOW_TEMPERATURE_ALARM",
				"resource":     "linecard1",
			},
			want: map[string]interface{}{
				"resource":      "linecard1",
				"alarm-type-id": "LOW_TEMPERATURE_ALARM",
				"severity":      "WARNING",
				"text":          "injected low temperature",
				"time-created":  wantTime,
				"status":        "RAISED",
			},
		},
		{
			name: "legacy empty-resource key with sharp",
			key:  "#HIGH_TEMPERATURE_ALARM",
			fields: map[string]string{
				"id":           "#HIGH_TEMPERATURE_ALARM",
				"time-created": "1792843160178746605",
				"text":         "high temperature",
				"severity":     "MAJOR",
				"type-id":      "HIGH_TEMPERATURE_ALARM",
				"resource":     "",
			},
			want: map[string]interface{}{
				"resource":      "",
				"alarm-type-id": "HIGH_TEMPERATURE_ALARM",
				"severity":      "MAJOR",
				"text":          "high temperature",
				"time-created":  wantTime,
				"status":        "RAISED",
			},
		},
		{
			name: "severity out of range maps to UNKNOWN",
			key:  "linecard1#X",
			fields: map[string]string{
				"id":           "linecard1#X",
				"time-created": "1792843160178746605",
				"text":         "t",
				"severity":     "99",
				"type-id":      "X",
				"resource":     "linecard1",
			},
			want: map[string]interface{}{
				"resource":      "linecard1",
				"alarm-type-id": "X",
				"severity":      "UNKNOWN",
				"text":          "t",
				"time-created":  wantTime,
				"status":        "RAISED",
			},
		},
		{
			name:    "missing time-created is an error",
			key:     "linecard1#Y",
			fields:  map[string]string{"id": "linecard1#Y", "text": "t", "severity": "MINOR", "type-id": "Y", "resource": "linecard1"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := TranslateCuralarmEntry(tc.fields, tc.key)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error, got nil: %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got  %v\nwant %v", got, tc.want)
			}
		})
	}
}
