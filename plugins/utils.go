package plugins

import (
	"regexp"
)

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
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
