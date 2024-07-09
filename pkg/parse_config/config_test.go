package parse_config

import (
	assert2 "github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestParseConfig(t *testing.T) {
	assert := assert2.New(t)
	c, err := ParseConfig()
	assert.NoError(err, "parse config return error")
	log.Printf("%+v", c)
}
