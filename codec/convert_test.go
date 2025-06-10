package codec

import (
	"reflect"
	"testing"

	"github.com/thinkgos/encoding/testdata/examplepb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func protoEqual(a, b any) bool {
	return proto.Equal(a.(proto.Message), b.(proto.Message))
}

type testStruct[T any] struct {
	name    string
	input   string
	output  T
	wantErr bool
}

func test_BuiltinType[T any](t *testing.T, tests []testStruct[T], checkFunc func(string) (T, error), equal func(any, any) bool) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, err := checkFunc(tt.input)
			switch {
			case err != nil && !tt.wantErr:
				t.Errorf("got unexpected error\n%#v", err)
			case err == nil && tt.wantErr:
				t.Errorf("did not error when expected")
			case !equal(ts, tt.output):
				t.Errorf(
					"when testing %s; got\n%#v\nexpected\n%#v",
					tt.name,
					ts,
					tt.output,
				)
			}
		})
	}
}
func test_BuiltinTypeSlices[T any](t *testing.T, tests []testStruct[T], checkFunc func(string, string) (T, error), equal func(any, any) bool) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, err := checkFunc(tt.input, ",")
			switch {
			case err != nil && !tt.wantErr:
				t.Errorf("got unexpected error\n%#v", err)
			case err == nil && tt.wantErr:
				t.Errorf("did not error when expected")
			case !equal(ts, tt.output):
				t.Errorf(
					"when testing %s; got\n%#v\nexpected\n%#v",
					tt.name,
					ts,
					tt.output,
				)
			}
		})
	}
}

func Test_String(t *testing.T) {
	tests1 := []testStruct[string]{
		{
			name:    "",
			input:   "hello world",
			output:  "hello world",
			wantErr: false,
		},
	}
	test_BuiltinType(t, tests1, String, reflect.DeepEqual)

	tests2 := []testStruct[*wrapperspb.StringValue]{
		{
			name:    "",
			input:   "hello world",
			output:  wrapperspb.String("hello world"),
			wantErr: false,
		},
	}
	test_BuiltinType(t, tests2, StringValue, reflect.DeepEqual)
}

func Test_StringSlice(t *testing.T) {
	tests := []testStruct[[]string]{
		{
			name:    "",
			input:   "hello,world",
			output:  []string{"hello", "world"},
			wantErr: false,
		},
	}
	test_BuiltinTypeSlices(t, tests, StringSlice, reflect.DeepEqual)
}

func Test_Bool(t *testing.T) {
	tests := []testStruct[*wrapperspb.BoolValue]{
		{
			name:    "true",
			input:   "true",
			output:  wrapperspb.Bool(true),
			wantErr: false,
		},
		{
			name:    "T",
			input:   "true",
			output:  wrapperspb.Bool(true),
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "11",
			output:  wrapperspb.Bool(false),
			wantErr: true,
		},
	}
	test_BuiltinType(t, tests, BoolValue, reflect.DeepEqual)
}

func Test_BoolSlice(t *testing.T) {
	tests := []testStruct[[]bool]{
		{
			name:    "true|false",
			input:   "true,false",
			output:  []bool{true, false},
			wantErr: false,
		},
		{
			name:    "T|f",
			input:   "T,f",
			output:  []bool{true, false},
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "11",
			output:  nil,
			wantErr: true,
		},
	}
	test_BuiltinTypeSlices(t, tests, BoolSlice, reflect.DeepEqual)
}

func Test_Float64(t *testing.T) {
	tests := []testStruct[*wrapperspb.DoubleValue]{
		{
			name:    "",
			input:   "1.0",
			output:  wrapperspb.Double(1.0),
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "1.x",
			output:  wrapperspb.Double(0),
			wantErr: true,
		},
	}
	test_BuiltinType(t, tests, DoubleValue, reflect.DeepEqual)
}

func Test_Float64Slice(t *testing.T) {
	tests := []testStruct[[]float64]{
		{
			name:    "",
			input:   "1.0,1.1",
			output:  []float64{1.0, 1.1},
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "1.x,x",
			output:  nil,
			wantErr: true,
		},
	}
	test_BuiltinTypeSlices(t, tests, Float64Slice, reflect.DeepEqual)
}

func Test_Float32(t *testing.T) {
	tests := []testStruct[*wrapperspb.FloatValue]{
		{
			name:    "",
			input:   "1.0",
			output:  wrapperspb.Float(1.0),
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "1.x",
			output:  wrapperspb.Float(0),
			wantErr: true,
		},
	}
	test_BuiltinType(t, tests, FloatValue, reflect.DeepEqual)
}

func Test_Float32Slice(t *testing.T) {
	tests := []testStruct[[]float32]{
		{
			name:    "",
			input:   "1.0,1.1",
			output:  []float32{1.0, 1.1},
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "1.x,x",
			output:  nil,
			wantErr: true,
		},
	}
	test_BuiltinTypeSlices(t, tests, Float32Slice, reflect.DeepEqual)
}

func Test_Int64(t *testing.T) {
	tests := []testStruct[*wrapperspb.Int64Value]{
		{
			name:    "",
			input:   "1",
			output:  wrapperspb.Int64(1),
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "1.x",
			output:  wrapperspb.Int64(0),
			wantErr: true,
		},
	}
	test_BuiltinType(t, tests, Int64Value, protoEqual)
}

func Test_Int64Slice(t *testing.T) {
	tests := []testStruct[[]int64]{
		{
			name:    "",
			input:   "1,2",
			output:  []int64{1, 2},
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "1.x,x",
			output:  nil,
			wantErr: true,
		},
	}
	test_BuiltinTypeSlices(t, tests, Int64Slice, reflect.DeepEqual)
}

func Test_Int32(t *testing.T) {
	tests := []testStruct[*wrapperspb.Int32Value]{
		{
			name:    "",
			input:   "1",
			output:  wrapperspb.Int32(1),
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "1.x",
			output:  wrapperspb.Int32(0),
			wantErr: true,
		},
	}
	test_BuiltinType(t, tests, Int32Value, reflect.DeepEqual)
}

func Test_Int32Slice(t *testing.T) {
	tests := []testStruct[[]int32]{
		{
			name:    "",
			input:   "1,2",
			output:  []int32{1, 2},
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "1.x,x",
			output:  nil,
			wantErr: true,
		},
	}
	test_BuiltinTypeSlices(t, tests, Int32Slice, reflect.DeepEqual)
}

func Test_Uint64(t *testing.T) {
	tests := []testStruct[*wrapperspb.UInt64Value]{
		{
			name:    "",
			input:   "1",
			output:  wrapperspb.UInt64(1),
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "1.x",
			output:  wrapperspb.UInt64(0),
			wantErr: true,
		},
	}
	test_BuiltinType(t, tests, UInt64Value, reflect.DeepEqual)
}

func Test_Uint64Slice(t *testing.T) {
	tests := []testStruct[[]uint64]{
		{
			name:    "",
			input:   "1,2",
			output:  []uint64{1, 2},
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "1.x,x",
			output:  nil,
			wantErr: true,
		},
	}
	test_BuiltinTypeSlices(t, tests, Uint64Slice, reflect.DeepEqual)
}

func Test_Uint32(t *testing.T) {
	tests := []testStruct[*wrapperspb.UInt32Value]{
		{
			name:    "",
			input:   "1",
			output:  wrapperspb.UInt32(1),
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "1.x",
			output:  wrapperspb.UInt32(0),
			wantErr: true,
		},
	}
	test_BuiltinType(t, tests, UInt32Value, reflect.DeepEqual)
}

func Test_Uint32Slice(t *testing.T) {
	tests := []testStruct[[]uint32]{
		{
			name:    "",
			input:   "1,2",
			output:  []uint32{1, 2},
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "1.x,x",
			output:  nil,
			wantErr: true,
		},
	}
	test_BuiltinTypeSlices(t, tests, Uint32Slice, reflect.DeepEqual)
}

func Test_Bytes(t *testing.T) {
	tests := []testStruct[*wrapperspb.BytesValue]{
		{
			name:    "base64 std",
			input:   "aGVsbG8=",
			output:  wrapperspb.Bytes([]byte("hello")),
			wantErr: false,
		},
		{
			name:    "base64 url",
			input:   "aGVsbG8_",
			output:  wrapperspb.Bytes([]byte("hello?")),
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "1.x,x",
			output:  wrapperspb.Bytes(nil),
			wantErr: true,
		},
	}
	test_BuiltinType(t, tests, BytesValue, reflect.DeepEqual)
}

func Test_BytesSlice(t *testing.T) {
	tests := []testStruct[[][]byte]{
		{
			name:    "",
			input:   "aGVsbG8=,aGVsbG8_",
			output:  [][]byte{[]byte("hello"), []byte("hello?")},
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "1.x,x",
			output:  nil,
			wantErr: true,
		},
	}
	test_BuiltinTypeSlices(t, tests, BytesSlice, reflect.DeepEqual)
}

func Test_Timestamp(t *testing.T) {
	tests := []testStruct[*timestamppb.Timestamp]{
		{
			name:  "a valid RFC3339 timestamp",
			input: `"2016-05-10T10:19:13.123Z"`,
			output: &timestamppb.Timestamp{
				Seconds: 1462875553,
				Nanos:   123000000,
			},
			wantErr: false,
		},
		{
			name:  "a valid RFC3339 timestamp without double quotation",
			input: "2016-05-10T10:19:13.123Z",
			output: &timestamppb.Timestamp{
				Seconds: 1462875553,
				Nanos:   123000000,
			},
			wantErr: false,
		},
		{
			name:    "invalid timestamp",
			input:   `"05-10-2016T10:19:13.123Z"`,
			output:  nil,
			wantErr: true,
		},
		{
			name:    "JSON number",
			input:   "123",
			output:  nil,
			wantErr: true,
		},
		{
			name:    "JSON bool",
			input:   "true",
			output:  nil,
			wantErr: true,
		},
	}
	test_BuiltinType(t, tests, Timestamp, protoEqual)
}

func Test_Duration(t *testing.T) {
	tests := []testStruct[*durationpb.Duration]{
		{
			name:  "a valid duration",
			input: `"123.456s"`,
			output: &durationpb.Duration{
				Seconds: 123,
				Nanos:   456000000,
			},
			wantErr: false,
		},
		{
			name:  "a valid duration without double quotation",
			input: "123.456s",
			output: &durationpb.Duration{
				Seconds: 123,
				Nanos:   456000000,
			},
			wantErr: false,
		},
		{
			name:    "invalid duration",
			input:   `"123years"`,
			output:  nil,
			wantErr: true,
		},
		{
			name:    "JSON number",
			input:   "123",
			output:  nil,
			wantErr: true,
		},
		{
			name:    "JSON bool",
			input:   "true",
			output:  nil,
			wantErr: true,
		},
	}
	test_BuiltinType(t, tests, Duration, protoEqual)
}

func Test_Enum(t *testing.T) {
	tests := []testStruct[int32]{
		{
			name:    "valid enum",
			input:   "woman",
			output:  1,
			wantErr: false,
		},
		{
			name:    "valid but num",
			input:   "1",
			output:  1,
			wantErr: false,
		},
		{
			name:    "invalid number",
			input:   "1.x",
			output:  0,
			wantErr: true,
		},
		{
			name:    "valid number buf invalid number",
			input:   "11",
			output:  0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, err := Enum(tt.input, examplepb.Sex_value)
			switch {
			case err != nil && !tt.wantErr:
				t.Errorf("got unexpected error\n%#v", err)
			case err == nil && tt.wantErr:
				t.Errorf("did not error when expected")
			case !reflect.DeepEqual(ts, tt.output):
				t.Errorf(
					"when testing %s; got\n%#v\nexpected\n%#v",
					tt.name,
					ts,
					tt.output,
				)
			}
		})
	}
}

func Test_EnumSlices(t *testing.T) {
	tests := []testStruct[[]int32]{
		{
			name:    "valid enum",
			input:   "woman,man",
			output:  []int32{1, 0},
			wantErr: false,
		},
		{
			name:    "valid but num",
			input:   "1,0",
			output:  []int32{1, 0},
			wantErr: false,
		},
		{
			name:    "invalid number",
			input:   "1.x",
			output:  nil,
			wantErr: true,
		},
		{
			name:    "valid number buf invalid number",
			input:   "11",
			output:  nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, err := EnumSlice(tt.input, ",", examplepb.Sex_value)
			switch {
			case err != nil && !tt.wantErr:
				t.Errorf("got unexpected error\n%#v", err)
			case err == nil && tt.wantErr:
				t.Errorf("did not error when expected")
			case !reflect.DeepEqual(ts, tt.output):
				t.Errorf(
					"when testing %s; got\n%#v\nexpected\n%#v",
					tt.name,
					ts,
					tt.output,
				)
			}
		})
	}
}
