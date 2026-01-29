package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsResolvableValue(t *testing.T) {
	assert.True(t, IsResolvableValue("env:PATH"))
	assert.True(t, IsResolvableValue("file:/config/app.txt//KeyName"))
	assert.True(t, IsResolvableValue("json:/config/app.json//database.host"))
	assert.True(t, IsResolvableValue("yaml:/config/app.yaml//server.port"))
	assert.True(t, IsResolvableValue("ini:/config/app.ini//Section.Key"))
	assert.False(t, IsResolvableValue("http://example.com"))
	assert.False(t, IsResolvableValue("https://example.com"))
	assert.False(t, IsResolvableValue("ftp://example.com"))
	assert.False(t, IsResolvableValue("gopher://example.com"))
}

func TestIsHostnameLike(t *testing.T) {
	assert.True(t, IsHostnameLike("example.com"))
	assert.True(t, IsHostnameLike("sub.domain.local"))
	assert.True(t, IsHostnameLike("localhost"))
	assert.True(t, IsHostnameLike("a-b.c"))

	assert.False(t, IsHostnameLike(""))
	assert.False(t, IsHostnameLike("-bad.example"))
	assert.False(t, IsHostnameLike("bad-.example"))
	assert.False(t, IsHostnameLike("bad..example"))
	assert.False(t, IsHostnameLike("example.com/"))
	assert.False(t, IsHostnameLike("example.com:80"))
	assert.False(t, IsHostnameLike("exa_mple.com"))
}

func TestIsAlphaNum(t *testing.T) {
	assert.True(t, isAlphaNum('a'))
	assert.True(t, isAlphaNum('A'))
	assert.True(t, isAlphaNum('0'))
	assert.True(t, isAlphaNum('9'))
	assert.False(t, isAlphaNum('-'))
	assert.False(t, isAlphaNum('_'))
	assert.False(t, isAlphaNum('!'))
	assert.False(t, isAlphaNum('$'))
	assert.False(t, isAlphaNum('%'))
	assert.False(t, isAlphaNum('*'))
	assert.False(t, isAlphaNum('/'))
	assert.False(t, isAlphaNum(':'))
	assert.False(t, isAlphaNum('@'))
	assert.False(t, isAlphaNum('['))
	assert.False(t, isAlphaNum('\\'))
	assert.False(t, isAlphaNum(']'))
	assert.False(t, isAlphaNum('^'))
}
