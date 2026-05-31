package onboarding

// renderInAppShell renders only the wizard center column for embedding in
// app.View() (FR-5.1). Progression is the header step pill only (FR-5.3).
func renderInAppShell(vm ViewModel) string {
	st := vm.Theme.Styles()
	w, h := vm.Width, vm.Height
	if vm.Model.Applied {
		return renderShellReady(vm, st, w, h)
	}
	return renderShellCenter(vm, st, w, h)
}
