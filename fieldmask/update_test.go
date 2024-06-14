package fieldmask

import (
	"testing"

	syntaxv1 "go.einride.tech/aip/proto/gen/einride/example/syntax/v1"
	"google.golang.org/genproto/googleapis/example/library/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestUpdate(t *testing.T) {
	t.Parallel()
	t.Run("should panic on different src and dst", func(t *testing.T) {
		t.Parallel()
		assert.Assert(t, cmp.Panics(func() {
			Update(&fieldmaskpb.FieldMask{}, &library.Book{}, &library.Shelf{})
		}))
	})

	t.Run("full replacement", func(t *testing.T) {
		t.Parallel()
		for _, tt := range []struct {
			name string
			src  proto.Message
			dst  proto.Message
		}{
			{
				name: "scalars",
				src: &syntaxv1.Message{
					Double:  111,
					Float:   111,
					Bool:    true,
					String_: "111",
					Bytes:   []byte{111},
				},
				dst: &syntaxv1.Message{
					Double:  222,
					Float:   222,
					Bool:    false,
					String_: "222",
					Bytes:   []byte{222},
				},
			},
			{
				name: "repeated",
				src: &syntaxv1.Message{
					RepeatedDouble: []float64{111},
					RepeatedFloat:  []float32{111},
					RepeatedBool:   []bool{true},
					RepeatedString: []string{"111"},
					RepeatedBytes:  [][]byte{{111}},
				},
				dst: &syntaxv1.Message{
					RepeatedDouble: []float64{222},
					RepeatedFloat:  []float32{222},
					RepeatedBool:   []bool{false},
					RepeatedString: []string{"222"},
					RepeatedBytes:  [][]byte{{222}},
				},
			},
			{
				name: "nested",
				src: &syntaxv1.Message{
					Message: &syntaxv1.Message{
						String_: "src",
					},
				},
				dst: &syntaxv1.Message{
					Message: &syntaxv1.Message{
						String_: "dst",
						Int64:   222,
					},
				},
			},
			{
				name: "maps",
				src: &syntaxv1.Message{
					MapStringString: map[string]string{
						"src-key": "src-value",
					},
					MapStringMessage: map[string]*syntaxv1.Message{
						"src-key": {
							String_: "src-value",
						},
					},
				},
				dst: &syntaxv1.Message{
					MapStringString: map[string]string{
						"dst-key": "dst-value",
					},
					MapStringMessage: map[string]*syntaxv1.Message{
						"dst-key": {
							String_: "dst-value",
						},
					},
				},
			},
			{
				name: "oneof: swap",
				src: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofString{
						OneofString: "src",
					},
				},
				dst: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage2{
						OneofMessage2: &syntaxv1.Message{
							String_: "dst",
						},
					},
				},
			},
			{
				name: "oneof: message swap",
				src: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage1{
						OneofMessage1: &syntaxv1.Message{
							String_: "src",
						},
					},
				},
				dst: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage2{
						OneofMessage2: &syntaxv1.Message{
							String_: "dst",
						},
					},
				},
			},
		} {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				srcClone := proto.Clone(tt.src)
				Update(&fieldmaskpb.FieldMask{Paths: []string{"*"}}, tt.dst, tt.src)
				assert.DeepEqual(t, srcClone, tt.dst, protocmp.Transform())
			})
		}
	})
	t.Run("wire set fields", func(t *testing.T) {
		t.Parallel()
		for _, tt := range []struct {
			name     string
			src      proto.Message
			dst      proto.Message
			expected proto.Message
		}{
			{
				name: "scalars",
				src: &syntaxv1.Message{
					Double: 111,
					Float:  111,
				},
				dst: &syntaxv1.Message{
					Double:  222,
					Float:   222,
					Bool:    false,
					String_: "222",
					Bytes:   []byte{222},
				},
				expected: &syntaxv1.Message{
					Double:  111,
					Float:   111,
					Bool:    false,
					String_: "222",
					Bytes:   []byte{222},
				},
			},
			{
				name: "repeated",
				src: &syntaxv1.Message{
					RepeatedDouble: []float64{111},
				},
				dst: &syntaxv1.Message{
					RepeatedDouble: []float64{222},
					RepeatedFloat:  []float32{222},
					RepeatedBool:   []bool{false},
					RepeatedString: []string{"222"},
					RepeatedBytes:  [][]byte{{222}},
				},
				expected: &syntaxv1.Message{
					RepeatedDouble: []float64{111},
					RepeatedFloat:  []float32{222},
					RepeatedBool:   []bool{false},
					RepeatedString: []string{"222"},
					RepeatedBytes:  [][]byte{{222}},
				},
			},
			{
				name: "nested",
				src: &syntaxv1.Message{
					Message: &syntaxv1.Message{String_: "src"},
				},
				dst: &syntaxv1.Message{
					Message: &syntaxv1.Message{
						String_: "dst",
						Int64:   222,
					},
				},
				expected: &syntaxv1.Message{
					Message: &syntaxv1.Message{
						String_: "src",
						Int64:   222,
					},
				},
			},
			{
				name: "nested: dst nil",
				src: &syntaxv1.Message{
					Message: &syntaxv1.Message{String_: "src"},
				},
				dst: &syntaxv1.Message{
					Message: nil,
				},
				expected: &syntaxv1.Message{
					Message: &syntaxv1.Message{String_: "src"},
				},
			},
			{
				name: "maps",
				src: &syntaxv1.Message{
					MapStringString: map[string]string{
						"src-key": "src-value",
					},
					MapStringMessage: map[string]*syntaxv1.Message{
						"src-key": {String_: "src-value"},
					},
				},
				dst: &syntaxv1.Message{
					MapStringString: map[string]string{
						"dst-key": "dst-value",
					},
					MapStringMessage: map[string]*syntaxv1.Message{
						"dst-key": {String_: "dst-value"},
					},
				},
				expected: &syntaxv1.Message{
					MapStringString: map[string]string{
						"src-key": "src-value",
					},
					MapStringMessage: map[string]*syntaxv1.Message{
						"src-key": {String_: "src-value"},
					},
				},
			},
			{
				name: "maps: dst nil",
				src: &syntaxv1.Message{
					MapStringString: map[string]string{
						"src-key": "src-value",
					},
					MapStringMessage: map[string]*syntaxv1.Message{
						"src-key": {String_: "src-value"},
					},
				},
				dst: &syntaxv1.Message{
					MapStringString:  nil,
					MapStringMessage: nil,
				},
				expected: &syntaxv1.Message{
					MapStringString: map[string]string{
						"src-key": "src-value",
					},
					MapStringMessage: map[string]*syntaxv1.Message{
						"src-key": {String_: "src-value"},
					},
				},
			},
			{
				name: "oneof",
				src: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage1{
						OneofMessage1: &syntaxv1.Message{String_: "src"},
					},
				},
				dst: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage1{
						OneofMessage1: &syntaxv1.Message{
							String_: "dst",
							Int64:   222,
						},
					},
				},
				expected: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage1{
						OneofMessage1: &syntaxv1.Message{
							String_: "src",
							Int64:   222,
						},
					},
				},
			},
			{
				name: "oneof: kind swap",
				src: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofString{
						OneofString: "src",
					},
				},
				dst: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage2{
						OneofMessage2: &syntaxv1.Message{
							String_: "dst",
						},
					},
				},
				expected: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofString{
						OneofString: "src",
					},
				},
			},
			{
				name: "oneof: message swap",
				src: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage1{
						OneofMessage1: &syntaxv1.Message{
							String_: "src",
						},
					},
				},
				dst: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage2{
						OneofMessage2: &syntaxv1.Message{
							String_: "dst",
						},
					},
				},
				expected: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage1{
						OneofMessage1: &syntaxv1.Message{
							String_: "src",
						},
					},
				},
			},
		} {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				Update(nil, tt.dst, tt.src)
				assert.DeepEqual(t, tt.expected, tt.dst, protocmp.Transform())
			})
		}
	})
	t.Run("paths", func(t *testing.T) {
		t.Parallel()
		for _, tt := range []struct {
			name     string
			paths    []string
			src      proto.Message
			dst      proto.Message
			expected proto.Message
		}{
			{
				name: "scalars",
				paths: []string{
					"double",
					"bytes",
				},
				src: &syntaxv1.Message{
					Double: 111,
					Float:  111,
					Bytes:  []byte{111},
				},
				dst: &syntaxv1.Message{
					Double:  222,
					Float:   222,
					Bool:    false,
					String_: "222",
					Bytes:   []byte{222},
				},
				expected: &syntaxv1.Message{
					Double:  111,
					Float:   222,
					Bytes:   []byte{111},
					Bool:    false,
					String_: "222",
				},
			},
			{
				name: "repeated scalar",
				paths: []string{
					"repeated_double",
					"repeated_string",
				},
				src: &syntaxv1.Message{
					RepeatedDouble: []float64{111},
					RepeatedFloat:  []float32{111},
				},
				dst: &syntaxv1.Message{
					RepeatedDouble: []float64{222},
					RepeatedString: []string{"222"},
					RepeatedBytes:  [][]byte{{222}},
				},
				expected: &syntaxv1.Message{
					RepeatedDouble: []float64{111},
					RepeatedBytes:  [][]byte{{222}},
				},
			},
			{
				name: "repeated message",
				paths: []string{
					"repeated_message",
				},
				src: &syntaxv1.Message{
					RepeatedMessage: []*syntaxv1.Message{
						{String_: "src"},
						{Int64: 111},
					},
				},
				dst: &syntaxv1.Message{
					RepeatedMessage: []*syntaxv1.Message{
						{Int64: 222},
						{String_: "dst"},
					},
				},
				expected: &syntaxv1.Message{
					RepeatedMessage: []*syntaxv1.Message{
						{String_: "src"},
						{Int64: 111},
					},
				},
			},
			{
				// can not update individual fields in a repeated message
				name: "repeated message: deep",
				paths: []string{
					"repeated_message.*.string",
				},
				src: &syntaxv1.Message{
					RepeatedMessage: []*syntaxv1.Message{
						{String_: "src"},
						{Int64: 111},
					},
				},
				dst: &syntaxv1.Message{
					RepeatedMessage: []*syntaxv1.Message{
						{Int64: 222},
						{String_: "dst"},
					},
				},
				expected: &syntaxv1.Message{
					RepeatedMessage: []*syntaxv1.Message{
						{Int64: 222},
						{String_: "dst"},
					},
				},
			},
			{
				name: "nested",
				paths: []string{
					"message",
				},
				src: &syntaxv1.Message{
					Message: &syntaxv1.Message{
						String_: "src",
					},
				},
				dst: &syntaxv1.Message{
					Message: &syntaxv1.Message{
						String_: "dst",
						Int64:   222,
					},
				},
				expected: &syntaxv1.Message{
					Message: &syntaxv1.Message{
						String_: "src",
					},
				},
			},
			{
				name: "nested: deep",
				paths: []string{
					"message.string",
				},
				src: &syntaxv1.Message{
					Message: &syntaxv1.Message{
						String_: "src",
					},
				},
				dst: &syntaxv1.Message{
					Message: &syntaxv1.Message{
						String_: "dst",
						Int64:   222,
					},
				},
				expected: &syntaxv1.Message{
					Message: &syntaxv1.Message{
						String_: "src",
						Int64:   222,
					},
				},
			},
			{
				name: "nested: dst nil",
				paths: []string{
					"message",
				},
				src: &syntaxv1.Message{
					Message: &syntaxv1.Message{
						String_: "src",
					},
				},
				dst: &syntaxv1.Message{
					Message: nil,
				},
				expected: &syntaxv1.Message{
					Message: &syntaxv1.Message{
						String_: "src",
					},
				},
			},
			{
				name: "nested: deep, dst nil",
				paths: []string{
					"message.string",
				},
				src: &syntaxv1.Message{
					Message: &syntaxv1.Message{
						String_: "src",
					},
				},
				dst: &syntaxv1.Message{
					Message: nil,
				},
				expected: &syntaxv1.Message{
					Message: &syntaxv1.Message{
						String_: "src",
					},
				},
			},
			{
				name: "nested: deep, src nil",
				paths: []string{
					"message.string",
				},
				src: &syntaxv1.Message{
					Message: nil,
				},
				dst: &syntaxv1.Message{
					Message: &syntaxv1.Message{
						String_: "src",
					},
				},
				expected: &syntaxv1.Message{
					Message: &syntaxv1.Message{},
				},
			},
			{
				name: "maps",
				paths: []string{
					"map_string_string",
				},
				src: &syntaxv1.Message{
					MapStringString: map[string]string{
						"src-key": "src-value",
					},
					MapStringMessage: map[string]*syntaxv1.Message{
						"src-key": {String_: "src-value"},
					},
				},
				dst: &syntaxv1.Message{
					MapStringString: map[string]string{
						"dst-key": "dst-value",
					},
					MapStringMessage: map[string]*syntaxv1.Message{
						"dst-key": {String_: "dst-value"},
					},
				},
				expected: &syntaxv1.Message{
					MapStringString: map[string]string{
						"src-key": "src-value",
					},
					MapStringMessage: map[string]*syntaxv1.Message{
						"dst-key": {String_: "dst-value"},
					},
				},
			},
			{
				// can not update individual entries in a map
				name: "maps: deep",
				paths: []string{
					"map_string_string.src1",
				},
				src: &syntaxv1.Message{
					MapStringString: map[string]string{
						"src1": "src1-value",
						"src2": "src2-value",
					},
				},
				dst: &syntaxv1.Message{
					MapStringString: map[string]string{
						"dst-key": "dst-value",
					},
				},
				expected: &syntaxv1.Message{
					MapStringString: map[string]string{
						"dst-key": "dst-value",
					},
				},
			},
			{
				name: "maps: dst nil",
				paths: []string{
					"map_string_string",
				},
				src: &syntaxv1.Message{
					MapStringString: map[string]string{
						"src-key": "src-value",
					},
					MapStringMessage: map[string]*syntaxv1.Message{
						"src-key": {String_: "src-value"},
					},
				},
				dst: &syntaxv1.Message{},
				expected: &syntaxv1.Message{
					MapStringString: map[string]string{
						"src-key": "src-value",
					},
				},
			},
			{
				name: "maps: src nil",
				paths: []string{
					"map_string_string",
				},
				src: &syntaxv1.Message{},
				dst: &syntaxv1.Message{
					MapStringString: map[string]string{
						"dst-key": "dst-value",
					},
					MapStringMessage: map[string]*syntaxv1.Message{
						"dst-key": {String_: "dst-value"},
					},
				},
				expected: &syntaxv1.Message{
					MapStringMessage: map[string]*syntaxv1.Message{
						"dst-key": {String_: "dst-value"},
					},
				},
			},
			{
				name: "oneof",
				paths: []string{
					"oneof_message1",
				},
				src: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage1{
						OneofMessage1: &syntaxv1.Message{
							String_: "src",
						},
					},
				},
				dst: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage1{
						OneofMessage1: &syntaxv1.Message{
							String_: "dst",
							Int64:   222,
						},
					},
				},
				expected: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage1{
						OneofMessage1: &syntaxv1.Message{
							String_: "src",
						},
					},
				},
			},
			{
				name: "oneof: kind swap",
				paths: []string{
					"oneof_string",
				},
				src: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofString{
						OneofString: "src",
					},
				},
				dst: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage2{
						OneofMessage2: &syntaxv1.Message{
							String_: "dst",
						},
					},
				},
				expected: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofString{
						OneofString: "src",
					},
				},
			},
			{
				name: "oneof: kind swap src nil",
				paths: []string{
					"oneof_message2",
				},
				src: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofString{
						OneofString: "src",
					},
				},
				dst: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage2{
						OneofMessage2: &syntaxv1.Message{
							String_: "dst",
						},
					},
				},
				expected: &syntaxv1.Message{
					Oneof: nil,
				},
			},
			{
				name: "oneof: deep",
				paths: []string{
					"oneof_message1.string",
				},
				src: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage1{
						OneofMessage1: &syntaxv1.Message{
							String_: "src",
						},
					},
				},
				dst: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage2{
						OneofMessage2: &syntaxv1.Message{
							String_: "dst",
						},
					},
				},
				expected: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage1{
						OneofMessage1: &syntaxv1.Message{
							String_: "src",
						},
					},
				},
			},
			{
				name: "oneof: deep src nil",
				paths: []string{
					"oneof_message2.string",
				},
				src: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage1{
						OneofMessage1: &syntaxv1.Message{
							String_: "src",
						},
					},
				},
				dst: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage2{
						OneofMessage2: &syntaxv1.Message{
							String_: "dst",
						},
					},
				},
				expected: &syntaxv1.Message{
					Oneof: &syntaxv1.Message_OneofMessage2{
						OneofMessage2: &syntaxv1.Message{},
					},
				},
			},
			{
				name: "message: src nil",
				paths: []string{
					"message",
				},
				src: &syntaxv1.Message{
					Message: nil,
				},
				dst: &syntaxv1.Message{
					Message: &syntaxv1.Message{
						Int32: 23,
					},
				},
				expected: &syntaxv1.Message{
					Message: nil,
				},
			},
		} {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				Update(&fieldmaskpb.FieldMask{Paths: tt.paths}, tt.dst, tt.src)
				assert.DeepEqual(t, tt.expected, tt.dst, protocmp.Transform())
			})
		}
	})
}
