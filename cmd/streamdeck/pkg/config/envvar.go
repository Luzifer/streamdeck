package config

import (
	"fmt"
	"os"
	"regexp"

	"go.yaml.in/yaml/v3"
)

var envVariablePattern = regexp.MustCompile(`\$\{env\.([A-Za-z_][A-Za-z0-9_]*)\}`)

func expandEnvVariables(node *yaml.Node) error {
	return expandEnvVariablesWithLookup(node, os.LookupEnv)
}

func expandEnvVariablesWithLookup(node *yaml.Node, lookup func(string) (string, bool)) error {
	return walkEnvVariables(node, lookup)
}

func replaceEnvVariables(value string, lookup func(string) (string, bool)) (string, error) {
	var err error

	expanded := envVariablePattern.ReplaceAllStringFunc(value, func(match string) string {
		if err != nil {
			return match
		}

		parts := envVariablePattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}

		replacement, ok := lookup(parts[1])
		if !ok {
			err = fmt.Errorf("environment variable %q is not set", parts[1])
			return match
		}

		return replacement
	})
	if err != nil {
		return "", err
	}

	return expanded, nil
}

func walkEnvVariables(node *yaml.Node, lookup func(string) (string, bool)) error {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case yaml.DocumentNode, yaml.SequenceNode:
		for _, child := range node.Content {
			if err := walkEnvVariables(child, lookup); err != nil {
				return err
			}
		}

	case yaml.MappingNode:
		for i := 1; i < len(node.Content); i += 2 {
			if err := walkEnvVariables(node.Content[i], lookup); err != nil {
				return err
			}
		}

	case yaml.ScalarNode:
		if node.Tag != "!!str" {
			return nil
		}

		expanded, err := replaceEnvVariables(node.Value, lookup)
		if err != nil {
			return err
		}

		node.Value = expanded
	}

	return nil
}
