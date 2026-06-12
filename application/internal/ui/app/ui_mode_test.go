package app

import "testing"

func TestUseRichLayout(t *testing.T) {
	t.Parallel()
	if useRichLayout("plain") || useRichLayout("json") {
		t.Fatal("plain/json must disable rich layout")
	}
	if !useRichLayout("auto") || !useRichLayout("rich") || !useRichLayout("") {
		t.Fatal("auto/rich/empty must enable rich layout")
	}
}
