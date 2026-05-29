package theme

var themes = map[string]Theme{
	"asagiri-dark": {
		Name: "asagiri-dark",
		Palette: Palette{
			Primary:    "#7AA2F7",
			Muted:      "#7F849C",
			Success:    "#9ECE6A",
			Warning:    "#E0AF68",
			Error:      "#F7768E",
			Border:     "#414868",
			Background: "#1A1B26",
		},
	},
	"asagiri-light": {
		Name: "asagiri-light",
		Palette: Palette{
			Primary:    "#005FB8",
			Muted:      "#5C6370",
			Success:    "#2B8A3E",
			Warning:    "#C77700",
			Error:      "#C92A2A",
			Border:     "#CED4DA",
			Background: "#F8F9FA",
		},
	},
	"minimal": {
		Name: "minimal",
		Palette: Palette{
			Primary:    "#FFFFFF",
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
			Primary:    "#00F5D4",
			Muted:      "#90E0EF",
			Success:    "#57CC99",
			Warning:    "#FFD166",
			Error:      "#FF4D6D",
			Border:     "#3A86FF",
			Background: "#0B132B",
		},
	},
}
