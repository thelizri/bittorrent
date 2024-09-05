package bencode

import (
	"fmt"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		data interface{}
		want string
	}{
		{"hello", "5:hello"},
		{52, "i52e"},
		{[]interface{}{"hello", 52}, "l5:helloi52ee"},
		{map[string]interface{}{"foo": "bar", "hello": 52}, "d3:foo3:bar5:helloi52ee"},
		{[]interface{}{[]interface{}{"spam", "eggs"}, []interface{}{"foo", 52}}, "ll4:spam4:eggsel3:fooi52eee"},
		{map[string]interface{}{"bar": "hello", "foo": map[string]interface{}{"baz": "qux", "hello": 52}}, "d3:bar5:hello3:food3:baz3:qux5:helloi52eee"},
	}

	for _, tt := range tests {
		have := Encode(tt.data)
		testname := fmt.Sprintf("%v", tt.data)

		t.Run(testname, func(t *testing.T) {
			want := tt.want

			if have != want {
				t.Errorf("have: %v, want %s", have, want)
			}
		})
	}
}
