package nmcslog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/invopop/jsonschema"
	"github.com/invopop/yaml"
	schemavalidate "github.com/qri-io/jsonschema"
)

func GenerateSchema(path string) (err error) {
	return generateSchema(path, false)
}
func GenerateSchemaWithComments(path string) (err error) {
	return generateSchema(path, true)
}

// Function to validate if all directories in the path exist, excluding the file name
func validatePath(path string) error {
	// Extract the directory portion of the path
	dirPath := filepath.Dir(path)

	// Clean the directory path to remove any trailing slashes
	cleanPath := filepath.Clean(dirPath)

	// Split the path into its directory components
	directories := strings.Split(cleanPath, string(os.PathSeparator))

	// Accumulate the path as we check each directory
	currentPath := ""
	if filepath.IsAbs(path) {
		currentPath = string(os.PathSeparator)
	}

	for _, dir := range directories {
		if dir == "" {
			continue
		}
		currentPath = filepath.Join(currentPath, dir)
		info, err := os.Stat(currentPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("directory does not exist: %s", currentPath)
			}
			return fmt.Errorf("error accessing directory %s: %v", currentPath, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("path component is not a directory: %s", currentPath)
		}
	}

	return nil
}

func generateSchema(path string, enableComments bool) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: generating configuration schema %q: %w", path, err)
		}
	}()

	if err := validatePath(path); err != nil {
		// path = "./config.schema.json"
		return fmt.Errorf("validating schema output path [%s]: %w", path, err)
	}

	r := new(jsonschema.Reflector)
	if enableComments {
		if err := r.AddGoComments("github.com/notmycloud/slog", "./"); err != nil {
			return fmt.Errorf("importing comments: %w", err)
		}
	}
	// Only tags explicitly marked as required instead of any that don't have `json:,omitempty`.
	r.RequiredFromJSONSchemaTags = true

	schema := r.Reflect(&Config{})

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling the schema to JSON: %w", err)
	}

	f, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer func() {
		if err := f.Close(); err != nil {
			// How can we handle this error correctly?
			// slog.Error("Closing Schema File", err)
			// panic(err)
		}
	}()

	if _, err := f.WriteString(string(data)); err != nil {
		return fmt.Errorf("writing config to [%s]: %w", path, err)
	}

	return nil
}

func ValidateSchema(path string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: generating configuration schema %q: %w", path, err)
		}
	}()

	if err := validatePath(path); err != nil {
		// path = "./config.schema.json"
		return fmt.Errorf("validating schema output path [%s]: %w", path, err)
	}

	r := new(jsonschema.Reflector)
	// Only tags explicitly marked as required instead of any that don't have `json:,omitempty`.
	r.RequiredFromJSONSchemaTags = true
	schema := r.Reflect(&Config{})
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("marshalling Schema JSON: %w", err)
	}

	configData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file [%s]: %w", path, err)
	}
	var configJSON []byte
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		configJSON, err = yaml.YAMLToJSON(configData)
		if err != nil {
			return fmt.Errorf("convert YAML config to JSON: %w", err)
		}
	case ".toml", ".tml":
		var conf Config
		_, err = toml.Decode(string(configData), &conf)
		if err != nil {
			return fmt.Errorf("decode TOML config: %w", err)
		}
		configJSON, err = json.Marshal(conf)
		if err != nil {
			return fmt.Errorf("convert TOML config to JSON: %w", err)
		}
	case ".json":
		configJSON = configData
	default:
		return fmt.Errorf("invalid file format [%s]", filepath.Ext(path))
	}

	validator := &schemavalidate.Schema{}
	if err := json.Unmarshal(schemaJSON, validator); err != nil {
		return fmt.Errorf("unmarshalling Schema JSON: %w", err)
	}
	var keyErrors []schemavalidate.KeyError
	keyErrors, err = validator.ValidateBytes(context.Background(), configJSON)
	if err != nil {
		return fmt.Errorf("validating configuration [%s]: %w", path, err)
	}
	var formattedErrors []error
	for i, keyError := range keyErrors {
		formattedErrors = append(formattedErrors, fmt.Errorf("key error [%d]: %v", i, keyError))
	}
	if len(keyErrors) > 0 {
		return fmt.Errorf("invalid configuration [%s]: %w", path, errors.Join(formattedErrors...))
	}

	return nil
}
