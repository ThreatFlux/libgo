package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// YAMLLoader implements Loader for YAML files.
type YAMLLoader struct {
	// Default config file path.
	DefaultPath string
}

// NewYAMLLoader creates a new YAML config loader.
func NewYAMLLoader(defaultPath string) *YAMLLoader {
	return &YAMLLoader{
		DefaultPath: defaultPath,
	}
}

// Load implements Loader.Load for YAML files.
func (l *YAMLLoader) Load(cfg *Config) error {
	if err := l.LoadFromFile(l.DefaultPath, cfg); err != nil {
		return fmt.Errorf("loading config from default path: %w", err)
	}

	if err := l.LoadWithOverrides(cfg); err != nil {
		return fmt.Errorf("applying environment overrides: %w", err)
	}

	return nil
}

// LoadFromFile implements Loader.LoadFromFile for YAML files.
func (l *YAMLLoader) LoadFromFile(filePath string, cfg *Config) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading config file %s: %w", filePath, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("unmarshaling YAML: %w", err)
	}

	return nil
}

// LoadWithOverrides implements Loader.LoadWithOverrides.
func (l *YAMLLoader) LoadWithOverrides(cfg *Config) error {
	return applyEnvironmentOverrides(cfg)
}

// applyEnvironmentOverrides applies environment variables as overrides to the config.
func applyEnvironmentOverrides(cfg *Config) error {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	// Process each field of the config struct.
	return walkStructForEnvOverrides(v, t, "")
}

// walkStructForEnvOverrides walks through a struct applying env var overrides.
func walkStructForEnvOverrides(v reflect.Value, t reflect.Type, prefix string) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Get the JSON tag (if any) to use as the env var name.
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		// Remove any options from the tag.
		tagParts := strings.Split(tag, ",")
		tag = tagParts[0]

		// Build the environment variable name.
		envName := buildEnvVarName(prefix, tag)

		// If this is a nested struct, recursively process it.
		if field.Type.Kind() == reflect.Struct {
			if err := walkStructForEnvOverrides(fieldValue, field.Type, envName); err != nil {
				return err
			}
			continue
		}

		// Look for an environment variable with this name.
		envValue, exists := os.LookupEnv(envName)
		if !exists {
			continue
		}

		// Apply the environment value to the field based on its type.
		if err := applyEnvValueToField(fieldValue, envValue); err != nil {
			return fmt.Errorf("applying env var %s: %w", envName, err)
		}
	}

	return nil
}

// buildEnvVarName constructs an environment variable name from prefix and field.
func buildEnvVarName(prefix, field string) string {
	parts := []string{}

	if prefix != "" {
		parts = append(parts, prefix)
	}

	parts = append(parts, field)

	// Join the parts and convert to uppercase.
	envName := strings.Join(parts, "_")
	return strings.ToUpper(envName)
}

// applyEnvValueToField sets a field's value from an environment variable string.
func applyEnvValueToField(fieldValue reflect.Value, envValue string) error {
	switch fieldValue.Kind() {
	case reflect.String:
		return applyStringValue(fieldValue, envValue)
	case reflect.Bool:
		return applyBoolValue(fieldValue, envValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return applyIntValue(fieldValue, envValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return applyUintValue(fieldValue, envValue)
	case reflect.Float32, reflect.Float64:
		return applyFloatValue(fieldValue, envValue)
	case reflect.Map:
		return applyMapValue(fieldValue, envValue)
	case reflect.Slice:
		return applySliceValue(fieldValue, envValue)
	default:
		return fmt.Errorf("unsupported field type: %s", fieldValue.Kind())
	}
}

// applyStringValue sets a string value.
func applyStringValue(fieldValue reflect.Value, envValue string) error {
	fieldValue.SetString(envValue)
	return nil
}

// applyBoolValue parses and sets a boolean value.
func applyBoolValue(fieldValue reflect.Value, envValue string) error {
	boolValue, err := strconv.ParseBool(envValue)
	if err != nil {
		return fmt.Errorf("parsing bool: %w", err)
	}
	fieldValue.SetBool(boolValue)
	return nil
}

// applyIntValue parses and sets an integer value, with special handling for duration.
func applyIntValue(fieldValue reflect.Value, envValue string) error {
	// Special handling for duration.
	if fieldValue.Type() == reflect.TypeOf(time.Duration(0)) {
		duration, err := time.ParseDuration(envValue)
		if err != nil {
			return fmt.Errorf("parsing duration: %w", err)
		}
		fieldValue.Set(reflect.ValueOf(duration))
		return nil
	}

	intValue, err := strconv.ParseInt(envValue, 10, 64)
	if err != nil {
		return fmt.Errorf("parsing int: %w", err)
	}
	fieldValue.SetInt(intValue)
	return nil
}

// applyUintValue parses and sets an unsigned integer value.
func applyUintValue(fieldValue reflect.Value, envValue string) error {
	uintValue, err := strconv.ParseUint(envValue, 10, 64)
	if err != nil {
		return fmt.Errorf("parsing uint: %w", err)
	}
	fieldValue.SetUint(uintValue)
	return nil
}

// applyFloatValue parses and sets a float value.
func applyFloatValue(fieldValue reflect.Value, envValue string) error {
	floatValue, err := strconv.ParseFloat(envValue, 64)
	if err != nil {
		return fmt.Errorf("parsing float: %w", err)
	}
	fieldValue.SetFloat(floatValue)
	return nil
}

// applyMapValue parses and sets a map value from comma-separated key:value pairs.
func applyMapValue(fieldValue reflect.Value, envValue string) error {
	// Maps in environment variables can be specified as key1:value1,key2:value2.
	// Merge with existing map instead of replacing it.
	if fieldValue.IsNil() {
		fieldValue.Set(reflect.MakeMap(fieldValue.Type()))
	}
	if envValue == "" {
		return nil
	}

	pairs := strings.Split(envValue, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) != 2 {
			return fmt.Errorf("invalid map format, expected key:value")
		}
		fieldValue.SetMapIndex(reflect.ValueOf(kv[0]), reflect.ValueOf(kv[1]))
	}
	return nil
}

// applySliceValue parses and sets a slice value from comma-separated values.
func applySliceValue(fieldValue reflect.Value, envValue string) error {
	// Slices in environment variables can be specified as value1,value2,value3.
	sliceType := fieldValue.Type().Elem()
	sliceValues := strings.Split(envValue, ",")

	sliceValue := reflect.MakeSlice(fieldValue.Type(), 0, len(sliceValues))

	for _, val := range sliceValues {
		elemValue, err := parseSliceElement(val, sliceType)
		if err != nil {
			return err
		}
		sliceValue = reflect.Append(sliceValue, elemValue)
	}

	fieldValue.Set(sliceValue)
	return nil
}

// parseSliceElement parses a single slice element value.
func parseSliceElement(val string, sliceType reflect.Type) (reflect.Value, error) {
	switch sliceType.Kind() {
	case reflect.String:
		return reflect.ValueOf(val), nil
	case reflect.Int, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("parsing slice int: %w", err)
		}
		return reflect.ValueOf(intVal).Convert(sliceType), nil
	// Add more types as needed.
	default:
		return reflect.Value{}, fmt.Errorf("unsupported slice element type: %s", sliceType.Kind())
	}
}
