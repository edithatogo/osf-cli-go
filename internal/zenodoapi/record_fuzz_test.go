package zenodoapi

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

// FuzzRecordNativePreservation exercises the source-preserving boundary, not a
// live provider. The corpus contains only fabricated records.
func FuzzRecordNativePreservation(f *testing.F) {
	for _, seed := range []string{
		`{"id":9007199254740993,"metadata":{"title":"synthetic"},"unknown":{"nested":[1,2]}}`,
		`{"id":"synthetic","files":{"entries":{"b":{"size":2},"a":{"size":1}},"order":["b"]}}`,
		`{"id":"synthetic","files":[{"key":"synthetic.txt","size":0}]}`,
		`null`, `{`, `[]`,
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64*1024 {
			t.Skip()
		}
		input := bytes.Clone(data)
		var first, second Record
		if err := json.Unmarshal(input, &first); err != nil {
			return
		}
		if err := json.Unmarshal(data, &second); err != nil || !reflect.DeepEqual(first, second) {
			t.Fatal("record decoding is nondeterministic")
		}
		want := bytes.TrimSpace(data)
		if !bytes.Equal(first.NativeJSON(), want) {
			t.Fatal("native JSON was changed or dropped")
		}
		// Neither a caller-owned result nor the input buffer may alias retained
		// evidence. Mutation here is confined to these synthetic fuzz inputs.
		copyOfNative := first.NativeJSON()
		if len(copyOfNative) != 0 {
			copyOfNative[0] ^= 0xff
		}
		if !bytes.Equal(first.NativeJSON(), want) {
			t.Fatal("NativeJSON leaked a mutable reference to retained evidence")
		}
		before := append([]byte(nil), want...)
		if len(input) != 0 {
			input[0] ^= 0xff
		}
		if !bytes.Equal(first.NativeJSON(), before) {
			t.Fatal("retained evidence aliases the input buffer")
		}
	})
}
