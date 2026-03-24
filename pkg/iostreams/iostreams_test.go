package iostreams

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSystem(t *testing.T) {
	ios := System()
	assert.NotNil(t, ios.In)
	assert.NotNil(t, ios.Out)
	assert.NotNil(t, ios.ErrOut)
}

func TestTest(t *testing.T) {
	ios, in, out, errOut := Test()
	assert.NotNil(t, ios)

	_, err := in.WriteString("input")
	assert.NoError(t, err)

	_, err = ios.Out.Write([]byte("output"))
	assert.NoError(t, err)
	assert.Equal(t, "output", out.String())

	_, err = ios.ErrOut.Write([]byte("error"))
	assert.NoError(t, err)
	assert.Equal(t, "error", errOut.String())
}

func TestIsTerminal_Buffer(t *testing.T) {
	ios, _, _, _ := Test()
	assert.False(t, ios.IsTerminal())
}
