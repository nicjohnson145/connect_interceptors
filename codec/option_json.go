// This is a copy-paste of connect's internal protoJSONCodec, with the only difference being that the marshal options
// are exposed as configuration params instead of only taking the defaults
package codec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"
)

func errNotProto(message any) error {
	if _, ok := message.(protoiface.MessageV1); ok {
		return fmt.Errorf("%T uses github.com/golang/protobuf, but connect-go only supports google.golang.org/protobuf: see https://go.dev/blog/protobuf-apiv2", message)
	}
	return fmt.Errorf("%T doesn't implement proto.Message", message)
}

type ProtoJSONCodecOpts struct {
	ProtoJsonOpts protojson.MarshalOptions
}

func NewProtoJSONCodec(opts ProtoJSONCodecOpts) protoJSONCodec {
	return protoJSONCodec{
		marshalOpts: opts.ProtoJsonOpts,
	}
}

type protoJSONCodec struct {
	marshalOpts protojson.MarshalOptions
}

var _ connect.Codec = (*protoJSONCodec)(nil)

func (c *protoJSONCodec) Name() string { return "json" }

func (c *protoJSONCodec) Marshal(message any) ([]byte, error) {
	protoMessage, ok := message.(proto.Message)
	if !ok {
		return nil, errNotProto(message)
	}
	return c.marshalOpts.Marshal(protoMessage)
}

func (c *protoJSONCodec) MarshalAppend(dst []byte, message any) ([]byte, error) {
	protoMessage, ok := message.(proto.Message)
	if !ok {
		return nil, errNotProto(message)
	}
	return c.marshalOpts.MarshalAppend(dst, protoMessage)
}

func (c *protoJSONCodec) Unmarshal(binary []byte, message any) error {
	protoMessage, ok := message.(proto.Message)
	if !ok {
		return errNotProto(message)
	}
	if len(binary) == 0 {
		return errors.New("zero-length payload is not a valid JSON object")
	}
	// Discard unknown fields so clients and servers aren't forced to always use
	// exactly the same version of the schema.
	options := protojson.UnmarshalOptions{DiscardUnknown: true}
	err := options.Unmarshal(binary, protoMessage)
	if err != nil {
		return fmt.Errorf("unmarshal into %T: %w", message, err)
	}
	return nil
}

func (c *protoJSONCodec) MarshalStable(message any) ([]byte, error) {
	// protojson does not offer a "deterministic" field ordering, but fields
	// are still ordered consistently by their index. However, protojson can
	// output inconsistent whitespace for some reason, therefore it is
	// suggested to use a formatter to ensure consistent formatting.
	// https://github.com/golang/protobuf/issues/1373
	messageJSON, err := c.Marshal(message)
	if err != nil {
		return nil, err
	}
	compactedJSON := bytes.NewBuffer(messageJSON[:0])
	if err = json.Compact(compactedJSON, messageJSON); err != nil {
		return nil, err
	}
	return compactedJSON.Bytes(), nil
}

func (c *protoJSONCodec) IsBinary() bool {
	return false
}
