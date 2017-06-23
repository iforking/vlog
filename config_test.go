package vlog

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"strings"
)

func TestLoadXmlConfig(t *testing.T) {
	root, err := LoadXmlConfig("vlog_sample.xml")
	assert.NoError(t, err)

	appenderElement := root.AppenderElements.AppenderElements[0]
	assert.Equal(t, "console", appenderElement.Name)
	assert.Equal(t, "ConsoleAppender", appenderElement.Type)
	assert.Equal(t, "default", appenderElement.TransformerName)

	transformerElement := root.TransformerElements.TransformerElements[0]
	assert.Equal(t, "default", transformerElement.Name)
	assert.Equal(t, "PatternTransformer", transformerElement.Type)
	assert.Equal(t, "<pattern>{time} [{level}] {logger} - {message}\\n</pattern>",
		strings.TrimSpace(string(transformerElement.InnerXML)))

	loggerElement := root.LoggerElements[0]
	assert.Equal(t, "github.com/user1", loggerElement.Name)
}

func TestPatternTransformerBuilder_Build(t *testing.T) {
	b := &PatternTransformerBuilder{}
	innerXml := []byte("<root><pattern>{time} [{level}] {logger} - {message}\\n</pattern></root>")
	tf, err := b.Build(innerXml)
	assert.NoError(t, err)
	assert.Equal(t, "{time} [{level}] {logger} - {message}\n", tf.(*PatternTransformer).pattern)
}
