package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MovieStoreGuy/retry/http/client"
)

func TestCreatingClient(t *testing.T) {
	t.Parallel()

	c, err := client.Default(100)
	assert.NoError(t, err, `Should not error with the default http client`)
	assert.NotNil(t, c)

	_, err = client.New(nil, 0)
	assert.Error(t, err, `Should fail to create with no provided http client`)

}
