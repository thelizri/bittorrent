package bencode

import (
	"fmt"
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	var tests = []struct {
		bencode string
		want    interface{}
	}{
		{"5:hello", "hello"},
		{"i52e", 52},
		{"l5:helloi52ee", []interface{}{"hello", 52}},
		{"d3:foo3:bar5:helloi52ee", map[string]interface{}{"foo": "bar", "hello": 52}},
		{"ll4:spam4:eggsel3:fooi52eee", []interface{}{[]interface{}{"spam", "eggs"}, []interface{}{"foo", 52}}},
		{"d3:bar5:hello3:food3:baz3:qux5:helloi52eee", map[string]interface{}{"bar": "hello", "foo": map[string]interface{}{"baz": "qux", "hello": 52}}},
	}

	var failingTests = []string{
		"4:hello",
		"5hello",
		"ihelloe",
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.bencode)
		t.Run(testname, func(t *testing.T) {
			have, err := Decode(tt.bencode)

			if err != nil {
				t.Error(err.Error())
			}

			switch want := tt.want.(type) {
			case map[string]interface{}:
				haveMap, ok := have.(map[string]interface{})

				if !ok {
					t.Errorf("have is not a map: %v", have)
				}

				if !compareInterfaceMaps(haveMap, want) {
					t.Errorf("have: %v, want: %v", haveMap, want)
				}
			case []interface{}:
				haveSlice, ok := have.([]interface{})

				if !ok {
					t.Errorf("have is not a slice: %v", have)
				}

				if !compareInterfaceSlices(haveSlice, want) {
					t.Errorf("have: %v, want %v", haveSlice, want)
				}
			default:
				if have != want {
					t.Errorf("have: %v, want %v", have, want)
				}
			}
		})
	}

	for _, bencode := range failingTests {
		_, err := Decode(bencode)
		testname := fmt.Sprintf("%s throws error", bencode)

		t.Run(testname, func(t *testing.T) {
			if err == nil {
				t.Error("Should have thrown an error but didn't")
			}
		})
	}
}

func compareInterfaceMaps(map1, map2 map[string]interface{}) bool {
	if len(map1) != len(map2) {
		return false
	}

	for key, value1 := range map1 {
		value2, ok := map2[key]
		if !ok {
			return false
		}

		switch v1 := value1.(type) {
		case map[string]interface{}:
			if v2, ok := value2.(map[string]interface{}); !ok {
				return false
			} else if !compareInterfaceMaps(v1, v2) {
				return false
			}
		case []interface{}:
			if v2, ok := value2.([]interface{}); !ok {
				return false
			} else if !compareInterfaceSlices(v1, v2) {
				return false
			}
		default:
			if value1 != value2 {
				return false
			}
		}
	}

	return true
}

func compareInterfaceSlices(slice1, slice2 []interface{}) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i := 0; i < len(slice1); i++ {
		v1 := reflect.ValueOf(slice1[i])
		v2 := reflect.ValueOf(slice2[i])

		if !v1.CanInterface() || !v2.CanInterface() {
			return false
		}

		if !reflect.DeepEqual(v1.Interface(), v2.Interface()) {
			return false
		}
	}

	return true
}
