package plugins

import (
	"regexp"

	"github.com/extemporalgenome/slug"
)

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ExtensionEnvironmentVariable struct {
	ProjectId             string `json:"projectId"`
	Key                   string `json:"key"`
	Type                  string `json:"type"`
	EnvironmentVariableId string `json:"environmentVariableId"`
}

func ConvertMapStringInterfaceToKV(formSpecMap map[string]interface{}, kv *[]KeyValue) error {
	for key, value := range formSpecMap {
		*kv = append(*kv, KeyValue{Key: key, Value: value.(string)})
	}
	return nil
}

func ConvertKVToMapStringInterface(kv []KeyValue, formSpecMap *map[string]interface{}) error {
	formMap := *formSpecMap

	for _, keyValue := range kv {
		tmpKv := keyValue
		formMap[tmpKv.Key] = &tmpKv.Value
	}

	formSpecMap = &formMap
	return nil
}

func ConvertKVToMapStringString(kv []KeyValue, formSpecMap *map[string]*string) error {

	formMap := *formSpecMap

	for _, keyValue := range kv {
		tmpKv := keyValue
		formMap[tmpKv.Key] = &tmpKv.Value
	}

	formSpecMap = &formMap

	return nil
}

func ConvertMapStringStringToKV(formSpecMap map[string]*string, kv *[]KeyValue) error {
	for key, value := range formSpecMap {
		*kv = append(*kv, KeyValue{Key: key, Value: *value})
	}
	return nil
}

func GetRegexParams(regEx, url string) (paramsMap map[string]string) {
	var compRegEx = regexp.MustCompile(regEx)
	match := compRegEx.FindStringSubmatch(url)

	paramsMap = make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return
}

func HstoreToMapStringString(hstore map[string]*string) map[string]string {
	mapStringString := map[string]string{}
	for k, v := range hstore {
		if k == "" || *v == "" {
			continue
		}
		mapStringString[k] = *v
	}

	return mapStringString
}

func MapStringStringToHstore(mapStringString map[string]string) map[string]*string {
	hstore := map[string]*string{}
	for k, v := range mapStringString {
		if k == "" || v == "" {
			continue
		}
		hstore[k] = &v
	}

	return hstore
}

func GetSlug(name string) string {
	return slug.Slug(name)
}
