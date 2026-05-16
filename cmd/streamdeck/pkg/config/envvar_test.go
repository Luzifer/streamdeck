package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

func TestExpandEnvVariables(t *testing.T) {
	t.Parallel()

	raw := []byte(`
plain: "prefix ${env.FOO} suffix"
multi: "${env.FOO}-${env.BAR}"
empty: "${env.EMPTY}"
number: 42
"${env.FOO}": "key stays literal"
`)

	var node yaml.Node
	require.NoError(t, yaml.Unmarshal(raw, &node))

	lookup := func(key string) (string, bool) {
		values := map[string]string{
			"BAR":   "bar",
			"EMPTY": "",
			"FOO":   "foo",
		}

		value, ok := values[key]
		return value, ok
	}

	require.NoError(t, expandEnvVariablesWithLookup(&node, lookup))

	var expanded map[string]any
	require.NoError(t, node.Decode(&expanded))

	assert.Equal(t, "prefix foo suffix", expanded["plain"])
	assert.Equal(t, "foo-bar", expanded["multi"])
	assert.Empty(t, expanded["empty"])
	assert.Equal(t, 42, expanded["number"])
	assert.Equal(t, "key stays literal", expanded["${env.FOO}"])
}

func TestExpandEnvVariablesErrorsOnMissingValue(t *testing.T) {
	t.Parallel()

	var node yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(`value: "${env.MISSING}"`), &node))

	err := expandEnvVariablesWithLookup(&node, func(string) (string, bool) {
		return "", false
	})
	require.Error(t, err)
}
