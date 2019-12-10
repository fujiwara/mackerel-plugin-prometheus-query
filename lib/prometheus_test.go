package promq

import (
	"testing"
	"time"

	model "github.com/prometheus/common/model"
)

func TestMetric(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2019-12-10T11:22:33+09:00")
	m := &metric{
		key:       "foo.bar",
		value:     123.456,
		timestamp: ts,
	}
	if m.String() != "foo.bar\t123.456\t1575944553" {
		t.Error("unexpected metric string", m.String())
	}
}

func TestFormatKey(t *testing.T) {
	mm := make(model.Metric)
	mm[model.LabelName("foo")] = model.LabelValue("FOO")
	mm[model.LabelName("bar")] = model.LabelValue("BAR")
	if s := formatKey(mm, "xxx.{foo}.{foo}.{bar}.{baz}.zzz"); s != "xxx.FOO.FOO.BAR._.zzz" {
		t.Error("unexpected formated string", s)
	}
}
