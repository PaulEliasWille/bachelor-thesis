package main

import (
	"encoding/json"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"strings"
)

func LoadJsonFromBytes(jsonBytes []byte) (interface{}, error) {
	var result interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func LoadJsonFromTomlBytes(tomlBytes []byte) (interface{}, error) {
	var result interface{}
	if err := toml.Unmarshal(tomlBytes, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func LoadJsonsFromYamlBytes(yamlBytes []byte) []interface{} {
	yamlString := string(yamlBytes)

	containedYamls := make([]string, 0)

	var sb strings.Builder
	for _, line := range strings.Split(yamlString, "\n") {
		if strings.TrimSpace(line) == "---" {
			currentYaml := strings.TrimSpace(sb.String())
			sb.Reset()

			if len(currentYaml) > 0 {
				containedYamls = append(containedYamls, currentYaml)
			}
		} else {
			sb.WriteString(fmt.Sprintf("%s\n", line))
		}
	}

	currentYaml := strings.TrimSpace(sb.String())
	if len(currentYaml) > 0 {
		containedYamls = append(containedYamls, currentYaml)
	}

	results := make([]interface{}, 0, len(containedYamls))
	for _, yamlString := range containedYamls {
		var result interface{}
		if err := yaml.Unmarshal([]byte(yamlString), &result); err != nil {
			continue
		}
		results = append(results, result)
	}

	return results
}

func LoadJsonFromFile(filename string) (interface{}, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return LoadJsonFromBytes(b)
}

func JsonResolve(obj interface{}, path []string) (interface{}, error) {
	if len(path) == 0 {
		return obj, nil
	}

	numWildcards := 0
	for _, key := range path {
		if key == "*" {
			numWildcards++
		}
	}

	if numWildcards > 1 {
		return nil, fmt.Errorf("only one wildcard is allowed in path %v", path)
	}

	switch obj := obj.(type) {
	case map[string]interface{}:
		key := path[0]
		if strings.HasPrefix(key, "{") && strings.HasSuffix(key, "}") {
			keys := strings.Split(key[1:len(key)-1], ",")
			for _, key := range keys {
				newPath := append([]string{key}, path[1:]...)
				result, err := JsonResolve(obj, newPath)
				if err == nil {
					return result, nil
				}
			}
			return nil, fmt.Errorf("cannot resolve path %v in object %v", path, obj)
		}
		value, ok := obj[key]
		if !ok {
			return nil, fmt.Errorf("cannot resolve path %v in object %v", path, obj)
		}
		return JsonResolve(value, path[1:])
	case []interface{}:
		key := path[0]
		if key == "*" {
			result := make([]interface{}, 0)
			for _, value := range obj {
				resolved, err := JsonResolve(value, path[1:])
				if err != nil {
					return nil, err
				}
				result = append(result, resolved)
			}
			return result, nil
		}

		index, err := strconv.Atoi(key)
		if err != nil {
			return nil, fmt.Errorf("cannot resolve path %v in object %v, index is no valid integer", path, obj)
		}

		if index < 0 || index >= len(obj) {
			return nil, fmt.Errorf("cannot resolve path %v in object %v, index out of range", path, obj)
		}

		value := obj[index]

		return JsonResolve(value, path)
	default:
		return nil, fmt.Errorf("cannot resolve path %v in object %v", path, obj)
	}
}

func JsonResolveString(obj interface{}, path []string) (string, error) {
	value, err := JsonResolve(obj, path)
	if err != nil {
		return "", err
	}

	valueString, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("value at path %v is not a string", path)
	}

	return valueString, nil
}

func JsonResolveFloat64(obj interface{}, path []string) (float64, error) {
	value, err := JsonResolve(obj, path)
	if err != nil {
		return 0, err
	}

	valueFloat, ok := value.(float64)
	if !ok {
		return 0, fmt.Errorf("value at path %v is not a number", path)
	}

	return valueFloat, nil
}

func JsonResolveInt64(obj interface{}, path []string) (int64, error) {
	valueFloat, err := JsonResolveFloat64(obj, path)
	if err != nil {
		return 0, err
	}

	return int64(valueFloat), nil
}

func JsonResolveArray(obj interface{}, path []string) ([]interface{}, error) {
	value, err := JsonResolve(obj, path)
	if err != nil {
		return nil, err
	}

	valueArray, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("value at path %v is not an array", path)
	}

	return valueArray, nil
}

func JsonResolveMap(obj interface{}, path []string) (map[string]interface{}, error) {
	value, err := JsonResolve(obj, path)
	if err != nil {
		return nil, err
	}

	valueMap, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("value at path %v is not a map", path)
	}

	return valueMap, nil
}
