package grpc

import (
	"bytes"
	"encoding/json"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"

	"google.golang.org/grpc/encoding"
)

const (
	codecJson  = "json"
	codecProto = "proto"
)

func init() {
	encoding.RegisterCodec(grpcJson{
		Marshaler: jsonpb.Marshaler{
			EmitDefaults: true,
			OrigName:     true,
		},
	})
}

type grpcJson struct {
	jsonpb.Marshaler
	jsonpb.Unmarshaler
}

// Name implements grpc encoding package Codec interface method,
// returns the name of the Codec implementation.
func (j grpcJson) Name() string {
	return codecJson
}

// Marshal implements grpc encoding package Codec interface method,returns the wire format of v.
func (j grpcJson) Marshal(v interface{}) (out []byte, err error) {
	if pm, ok := v.(proto.Message); ok {
		b := new(bytes.Buffer)
		err := j.Marshaler.Marshal(b, pm)
		if err != nil {
			return nil, err
		}
		return b.Bytes(), nil
	}
	return json.Marshal(v)
}

// Unmarshal implements grpc encoding package Codec interface method,Unmarshal parses the wire format into v.
func (j grpcJson) Unmarshal(data []byte, v interface{}) (err error) {
	if pm, ok := v.(proto.Message); ok {
		b := bytes.NewBuffer(data)
		return j.Unmarshaler.Unmarshal(b, pm)
	}
	return json.Unmarshal(data, v)
}
