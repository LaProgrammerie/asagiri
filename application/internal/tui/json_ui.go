package tui

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type jsonUI struct {
	out io.Writer
	enc *json.Encoder
}

func newJSONUI(out io.Writer) *jsonUI {
	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)
	return &jsonUI{out: out, enc: enc}
}

func (j *jsonUI) Box(title, body string) {
	_ = j.enc.Encode(map[string]any{"ts": time.Now().UTC(), "event": "box", "title": title, "body": body})
}

func (j *jsonUI) ProgressLine(label string, pct float64) {
	_ = j.enc.Encode(map[string]any{"ts": time.Now().UTC(), "event": "progress", "label": label, "pct": pct})
}

func (j *jsonUI) Event(kind string, payload any) {
	_ = j.enc.Encode(map[string]any{"ts": time.Now().UTC(), "event": kind, "data": payload})
}

func (j *jsonUI) Printf(format string, args ...any) {
	j.Event("log", fmt.Sprintf(format, args...))
}
