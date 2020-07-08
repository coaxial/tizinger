package credentials

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTidal(t *testing.T) {
	want := []TidalAccount{{Username: "mockuser@example.org", Password: "secret"}}
	credentialsFile = "../../fixtures/credentials/mock-credentials.yaml"
	defer func() { credentialsFile = "credentials.yml" }()

	got, err := Tidal()

	assert.Nil(t, err, "shouldn't have errored")
	assert.Equal(t, want, got, "should return the Tidal credentials")
}
