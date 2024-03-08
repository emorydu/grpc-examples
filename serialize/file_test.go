package serialize_test

import (
	"testing"

	"ed.io/grpc-examples/pb"
	"ed.io/grpc-examples/sample"
	"ed.io/grpc-examples/serialize"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestFileSerializer(t *testing.T) {
	t.Parallel()

	binaryFile := "../tmp/laptop.bin"
	jsonFile := "../tmp/laptop.json"

	laptop1 := sample.NewLaptop()

	err := serialize.WriteProtobufToBinaryFile(laptop1, binaryFile)
	require.NoError(t, err)

	laptop2 := &pb.Laptop{}

	serialize.ReadProtobufFromBinaryFile(binaryFile, laptop2)
	require.NoError(t, err)
	require.True(t, proto.Equal(laptop1, laptop2))

	err = serialize.WriteProtobufToJSONFile(laptop1, jsonFile)
	require.NoError(t, err)
}
