package main

import (
	pb "github.com/redhatinsights/yggdrasil/protocol"
	"testing"
)

func TestDispatch(t *testing.T) {
	input := &pb.Data{}

	got := jsonData(input)
	want := "{\"response_to\":\"\",\"metadata\":null,\"content\":null,\"directive\":\"\"}"

	if string(got) != want {
		t.Fatalf(`Got: %q, Wanted: %q`, string(got), want)
	}
}
