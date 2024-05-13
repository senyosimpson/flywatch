package flywatch

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	var config Config
	_, err := toml.DecodeFile("../example.toml", &config)
	if err != nil {
		t.Fatalf("error decoding config: %s", err.Error())
	}

	assert.Equal(t, config.Name, "hellofly-deployment")
	assert.Equal(t, config.App, "senyo-hellofly")
	assert.Equal(t, config.Build.Image, "flyio/hellofly:latest")

	assert.Equal(t, config.Compute[0].MemoryMB, 1024)
	assert.Equal(t, config.Compute[0].CPUs, 2)
	assert.Equal(t, config.Compute[0].CPUKind, "shared")

	assert.Equal(t, config.Replicas[0].Region, "jnb")
	assert.Equal(t, config.Replicas[0].Count, 2)
	assert.Equal(t, config.Replicas[1].Region, "ams")
	assert.Equal(t, config.Replicas[1].Count, 3)
}
