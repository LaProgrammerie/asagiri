package components

import (
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
)

// Panel renders a titled container.
func Panel(title, body string, th theme.Theme) string {
	st := th.Styles()
	header := st.PanelTitle.Render(title)
	content := st.Fg.Render(body)
	return st.Panel.Render(header + "\n" + content)
}
