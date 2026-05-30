package config_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func TestIsTemplateDefaultProjectName(t *testing.T) {
	require.True(t, config.IsTemplateDefaultProjectName(""))
	require.True(t, config.IsTemplateDefaultProjectName("my-project"))
	require.False(t, config.IsTemplateDefaultProjectName("chatbot"))
}

func TestIsTemplateDefaultBranchPrefix(t *testing.T) {
	require.True(t, config.IsTemplateDefaultBranchPrefix("asagiri"))
	require.False(t, config.IsTemplateDefaultBranchPrefix("chatbot"))
}

func TestIsTemplateDefaultValidationCommands(t *testing.T) {
	require.True(t, config.IsTemplateDefaultValidationCommands(nil))
	require.True(t, config.IsTemplateDefaultValidationCommands(config.DefaultGoValidationCommands("x")))
	require.False(t, config.IsTemplateDefaultValidationCommands([]config.ValidationCommand{
		{Name: "test", Command: "castor qa:phpunit", Required: true},
	}))
}
