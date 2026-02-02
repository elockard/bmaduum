package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand(t *testing.T) {
	cmd := newVersionCommand()

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "bmad-automate version")
	assert.Contains(t, output, Version)
	assert.True(t, strings.Contains(output, "1.1.0") || strings.Contains(output, "."))
}

func TestVersionCommand_NoArgs(t *testing.T) {
	cmd := newVersionCommand()

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Version command should work with no args
	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "bmad-automate version")
}
