package tracer

import (
	"encoding/json"
	"reflect"

	"github.com/epsagon/epsagon-go/protocol"
)

const maskedValue = "****"

func arrayToHitMap(arr []string) map[string]bool {
	hitMap := make(map[string]bool)
	for _, k := range arr {
		hitMap[k] = true
	}
	return hitMap
}

func maskNestedJSONKeys(decodedJSON interface{}, ignoredKeysMap map[string]bool) (interface{}, bool) {
	var changed bool
	decodedValue := reflect.ValueOf(decodedJSON)
	if decodedValue.Kind() == reflect.Invalid || decodedValue.IsZero() {
		return decodedJSON, false
	}
	switch decodedValue.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < decodedValue.Len(); i++ {
			nestedValue := decodedValue.Index(i)
			newNestedValue, indexChanged := maskNestedJSONKeys(nestedValue.Interface(), ignoredKeysMap)
			if indexChanged {
				nestedValue.Set(reflect.ValueOf(newNestedValue))
				changed = true
			}
		}
	case reflect.Map:
		for _, key := range decodedValue.MapKeys() {
			if ignoredKeysMap[key.String()] {
				decodedValue.SetMapIndex(key, reflect.ValueOf(maskedValue))
				changed = true
			} else {
				nestedValue := decodedValue.MapIndex(key)
				newNestedValue, valueChanged := maskNestedJSONKeys(nestedValue.Interface(), ignoredKeysMap)
				if valueChanged {
					decodedValue.SetMapIndex(key, reflect.ValueOf(newNestedValue))
					changed = true
				}
			}
		}
	}
	return decodedValue.Interface(), changed
}

// maskIgnoredKeys masks all the keys in the
// event resource metadata that are in ignoredKeys, swapping them with '****'.
// Metadata values that are json decodable will have their nested keys masked as well.
func (tracer *epsagonTracer) maskEventIgnoredKeys(event *protocol.Event, ignoredKeys []string) {
	ignoredKeysMap := arrayToHitMap(ignoredKeys)
	for key, value := range event.Resource.Metadata {
		if ignoredKeysMap[key] {
			event.Resource.Metadata[key] = maskedValue
		} else {
			var decodedJSON interface{}
			err := json.Unmarshal([]byte(value), &decodedJSON)
			if err == nil {
				newValue, changed := maskNestedJSONKeys(decodedJSON, ignoredKeysMap)
				if changed {
					encodedNewValue, err := json.Marshal(newValue)
					if err == nil {
						event.Resource.Metadata[key] = string(encodedNewValue)
					} else {
						exception := createException("internal json encode error", err.Error())
						if tracer.Stopped() {
							tracer.exceptions = append(tracer.exceptions, exception)
						} else {
							tracer.AddException(exception)
						}
					}
				}
			}
		}
	}
}
