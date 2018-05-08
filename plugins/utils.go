package plugins

import (
	"math/rand"
	"regexp"

	"github.com/extemporalgenome/slug"
)

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

func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
