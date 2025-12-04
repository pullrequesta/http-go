package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultServerOptions(t *testing.T) {
	var hf Handler

	testCases := []struct {
		name string
		opt  ServerOption
	}{
		{
			name: "tcp",
			opt:  WithTCP(),
		},
		{
			name: "udp",
			opt:  WithUDP(),
		},
	}

	for _, tc := range testCases {
		srv := NewServer(tc.opt)
		err := srv.Serve(hf)
		assert.NoError(t, err)
		err2 := srv.Close()
		assert.NoError(t, err2)
	}

}
