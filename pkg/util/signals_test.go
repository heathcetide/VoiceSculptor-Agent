package util

import (
	constants "VoiceSculptor/pkg/constant"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignal(t *testing.T) {
	var val string
	var eid uint
	eid = Sig().Connect("mock_test", func(sender any, params ...any) {
		val = sender.(string)
		assert.True(t, Sig().inLoop)
		Sig().Disconnect("mock_test", eid)
	})
	Sig().Emit("mock_test", "unittest")
	assert.Equal(t, val, "unittest")
	assert.Equal(t, 0, len(Sig().events))
	Sig().Clear("mock_test", constants.SigUserResetPassword, constants.SigUserVerifyEmail)
	assert.Equal(t, 0, len(Sig().sigHandlers))
}
