package config

import (
	"bytes"
	"fmt"

	"go.yaml.in/yaml/v3"
)

type (
	// DynamicAttributes stores raw YAML attributes until a module decodes them.
	DynamicAttributes = yaml.Node
)

// DecodeAttributes decodes raw dynamic attributes into a typed attribute struct.
func DecodeAttributes[T any](node DynamicAttributes) (t T, err error) {
	buf := new(bytes.Buffer)
	if err = yaml.NewEncoder(buf).Encode(node); err != nil {
		return t, fmt.Errorf("encoding attributes: %w", err)
	}

	decoder := yaml.NewDecoder(buf)
	decoder.KnownFields(true)

	if err = decoder.Decode(&t); err != nil {
		return t, fmt.Errorf("decoding attributes: %w", err)
	}

	return t, nil
}

// EncodeAttributes encodes typed attributes into the dynamic attribute format.
func EncodeAttributes(v any) (da DynamicAttributes, err error) {
	if err = da.Encode(v); err != nil {
		return da, fmt.Errorf("encoding attributes: %w", err)
	}

	return da, nil
}
