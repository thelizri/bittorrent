package bencode

import (
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"
)

func encodeString(value string) string {
	return fmt.Sprintf("%d:%v", len(value), value)
}

func encodeInt(value int) string {
	return fmt.Sprintf("i%ve", value)
}

func encodeList(list []interface{}) string {
	result := ""
	for i := range list {
		result += Encode(list[i])
	}

	return fmt.Sprintf("l%ve", result)
}

func encodeDict(dict map[string]interface{}) string {
	result := ""
	keys := make([]string, 0, len(dict))
	for key := range dict {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		result += encodeString(key)
		result += Encode(dict[key])
	}

	return fmt.Sprintf("d%ve", result)
}

func Encode(value interface{}) string {
	switch v := value.(type) {
	case string:
		return encodeString(v)
	case int:
		return encodeInt(v)
	case []interface{}:
		return encodeList(v)
	case map[string]interface{}:
		return encodeDict(v)
	default:
		log.Fatal("error encoding value")
	}

	return ""
}
