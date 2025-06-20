package proto

import (
	"bytes"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/thinkgos/encoding/testdata/examplepb"
)

type testInvalidProtoMessage struct {
	Id int64
}

var message = &examplepb.ABitOfEverything{
	SingleNested:        &examplepb.ABitOfEverything_Nested{},
	RepeatedStringValue: nil,
	MappedStringValue:   nil,
	MappedNestedValue:   nil,
	RepeatedEnumValue:   nil,
	TimestampValue:      &timestamppb.Timestamp{},
	Uuid:                "6EC2446F-7E89-4127-B3E6-5C05E6BECBA7",
	Nested: []*examplepb.ABitOfEverything_Nested{
		{
			Name:   "foo",
			Amount: 12345,
		},
	},
	Uint64Value: 0xFFFFFFFFFFFFFFFF,
	EnumValue:   examplepb.NumericEnum_ONE,
	OneofValue: &examplepb.ABitOfEverything_OneofString{
		OneofString: "bar",
	},
	MapValue: map[string]examplepb.NumericEnum{
		"a": examplepb.NumericEnum_ONE,
		"b": examplepb.NumericEnum_ZERO,
	},
}

func TestCodec_ContentType(t *testing.T) {
	var m Codec

	want := "application/x-protobuf"
	if got := m.ContentType(struct{}{}); got != want {
		t.Errorf("m.ContentType(_) failed, got = %q; want %q; ", got, want)
	}
}

func TestCodec_MarshalUnmarshal(t *testing.T) {
	m := Codec{}

	// Marshal
	buffer, err := m.Marshal(message)
	if err != nil {
		t.Fatalf("Marshalling returned error: %s", err.Error())
	}
	// Unmarshal
	unmarshalled := &examplepb.ABitOfEverything{}
	err = m.Unmarshal(buffer, unmarshalled)
	if err != nil {
		t.Fatalf("Unmarshalling returned error: %s", err.Error())
	}
	if !proto.Equal(unmarshalled, message) {
		t.Errorf(
			"Unmarshalled didn't match original message: (original = %v) != (unmarshalled = %v)",
			unmarshalled,
			message,
		)
	}

	// invalid proto message
	// Marshal
	_, err = m.Marshal(&testInvalidProtoMessage{Id: 11})
	if err == nil {
		t.Fatalf("Marshal should returned an error")
	}
	// Unmarshal
	unmarshalled1 := &testInvalidProtoMessage{}
	err = m.Unmarshal(buffer, unmarshalled1)
	if err == nil {
		t.Fatalf("Unmarshal should returned an error")
	}
}

func TestCodec_EncoderDecoder(t *testing.T) {
	m := Codec{}

	var buf bytes.Buffer

	encoder := m.NewEncoder(&buf)
	decoder := m.NewDecoder(&buf)

	// Encode
	err := encoder.Encode(message)
	if err != nil {
		t.Fatalf("Encoding returned error: %s", err.Error())
	}
	// Decode
	unencoded := &examplepb.ABitOfEverything{}
	err = decoder.Decode(unencoded)
	if err != nil {
		t.Fatalf("Unmarshalling returned error: %s", err.Error())
	}
	if !proto.Equal(unencoded, message) {
		t.Errorf(
			"Unencoded didn't match original message: (original = %v) != (unencoded = %v)",
			unencoded,
			message,
		)
	}

	// invalid proto message
	// Encode
	err = encoder.Encode(&testInvalidProtoMessage{Id: 11})
	if err == nil {
		t.Fatalf("Encode should returned an error")
	}
	// Decode
	unmarshalled1 := &testInvalidProtoMessage{}
	err = decoder.Decode(unmarshalled1)
	if err == nil {
		t.Fatalf("Decode should returned an error")
	}
}
