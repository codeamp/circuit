package transistor

import (
	"encoding/json"
	"errors"
	"math/rand"
	"reflect"
	"regexp"
	"time"

	log "github.com/codeamp/logger"
)

func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func SliceContains(name string, list []string) bool {
	for _, b := range list {
		matched, err := regexp.MatchString(b, name)
		if err != nil {
			log.InfoWithFields("SliceContains method encountered an error", log.Fields{
				"regex":  b,
				"string": name,
				"error":  err,
			})
		}

		if matched {
			return true
		}
	}

	// Moved outside of loop as this would return a debug log for every string that doesn't match
	// regardless of if we found the match in the haystack or not.
	// This way it only prints a debug if the regex didn't match ALL of the candidates
	// ADB
	log.DebugWithFields("SliceContains regex not matched", log.Fields{
		"string": name,
		"list":   list,
	})

	return false
}

func MapPayload(name string, event *Event) error {
	if typ, ok := EventRegistry[name]; ok {
		d, _ := json.Marshal(event.Payload)
		val := reflect.New(reflect.TypeOf(typ))
		json.Unmarshal(d, val.Interface())
		event.Payload = val.Elem().Interface()
		return nil
	}
	return errors.New("no match")
}
