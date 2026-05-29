package components

// VirtualWindow selects a visible slice from a long list.
type VirtualWindow struct {
	Offset int
	Limit  int
	Total  int
}

// SliceWindow returns offset/limit for virtual scrolling.
func SliceWindow(total, cursor, visible int) VirtualWindow {
	if visible <= 0 {
		visible = 10
	}
	if total <= 0 {
		return VirtualWindow{Total: 0, Limit: visible}
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= total {
		cursor = total - 1
	}
	offset := cursor - visible/2
	if offset < 0 {
		offset = 0
	}
	if offset+visible > total {
		offset = total - visible
		if offset < 0 {
			offset = 0
		}
	}
	limit := visible
	if offset+limit > total {
		limit = total - offset
	}
	return VirtualWindow{Offset: offset, Limit: limit, Total: total}
}

// VisibleSlice returns the sub-slice to render.
func VisibleSlice[T any](items []T, w VirtualWindow) []T {
	if len(items) == 0 || w.Limit <= 0 {
		return nil
	}
	start := w.Offset
	if start >= len(items) {
		return nil
	}
	end := start + w.Limit
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}
