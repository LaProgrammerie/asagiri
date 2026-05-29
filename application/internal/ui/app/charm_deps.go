package app

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"
)

var (
	_ = help.Model{}
	_ = glamour.WithAutoStyle
	_ = huh.NewForm
)
