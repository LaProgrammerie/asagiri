package theme

var themes = map[string]Theme{
	"asagiri-dark": {
		Name: "asagiri-dark",
		Palette: Palette{
			Primary:    "#9B6DFF",
			Foreground: "#E5E7EB",
			Muted:      "#6B7280",
			Success:    "#2DD4BF",
			Warning:    "#F59E0B",
			Error:      "#F87171",
			Border:     "#3D3552",
			Background: "#0D0F14",
			Surface:    "#141820",
		},
	},
	"asagiri-light": {
		Name: "asagiri-light",
		Palette: Palette{
			Primary:    "#6D28D9",
			Foreground: "#0F172A",
			Muted:      "#64748B",
			Success:    "#0D9488",
			Warning:    "#D97706",
			Error:      "#DC2626",
			Border:     "#CBD5E1",
			Background: "#F8F9FB",
			Surface:    "#EEF2F7",
		},
	},
	"minimal": {
		Name: "minimal",
		Palette: Palette{
			Primary:    "#FFFFFF",
			Foreground: "#FFFFFF",
			Muted:      "#C0C0C0",
			Success:    "#B5E48C",
			Warning:    "#FFD166",
			Error:      "#EF476F",
			Border:     "#7A7A7A",
			Background: "#000000",
		},
	},
	"high-contrast": {
		Name: "high-contrast",
		Palette: Palette{
			Primary:    "#00FFFF",
			Foreground: "#FFFFFF",
			Muted:      "#FFFFFF",
			Success:    "#00FF00",
			Warning:    "#FFFF00",
			Error:      "#FF0000",
			Border:     "#FFFFFF",
			Background: "#000000",
		},
	},
	"cyber": {
		Name: "cyber",
		Palette: Palette{
			Primary:    "#00F5FF",
			Foreground: "#E0FBFC",
			Muted:      "#90E0EF",
			Success:    "#39FF14",
			Warning:    "#FFD166",
			Error:      "#FF4D6D",
			Border:     "#3A86FF",
			Background: "#0B132B",
		},
	},
}
