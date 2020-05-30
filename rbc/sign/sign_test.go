package sign

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var testPubKey PublicKey
var testPrivateKey PrivateKey
var testWrongKey PrivateKey
const testMessage = "Test"

func init(){
	testPubKey, testPrivateKey = GenerateKey()
	_, testWrongKey = GenerateKey()
}

func TestVerify(t *testing.T){
	signedMessage := Sign(testPrivateKey, []byte(testMessage))
	wrongSignedMessage := Sign(testWrongKey, []byte(testMessage))

	decode, verified := Verify(testPubKey, signedMessage)
	assert.True(t, verified)
	assert.Equal(t, testMessage, string(decode))

	_, verified = Verify(testPubKey, wrongSignedMessage)
	assert.False(t, verified)
}
