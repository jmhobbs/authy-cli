package store_test

import (
	"os"
	"testing"

	"github.com/jmhobbs/authy-cli/model"
	"github.com/jmhobbs/authy-cli/store"
	"github.com/stretchr/testify/assert"
)

func TestStoreReadWrite(t *testing.T) {
	dir, err := os.MkdirTemp("", "store")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	s, err := store.New(dir, func() (string, error) { return "top secret", nil })
	assert.NoError(t, err)

	expectedConfig := store.Config{
		AuthyId: 12345,
		Device: store.Device{
			Id:         54321,
			SecretSeed: "secret-seed",
			ApiKey:     "api-key",
		},
	}

	err = s.WriteConfig(expectedConfig)
	assert.Nil(t, err)

	actualConfig, err := s.Config()
	assert.Nil(t, err)

	assert.Equal(t, &expectedConfig, actualConfig)

	expectedApps := []model.App{
		{
			Id:      "app-one",
			Name:    "Sendgrid",
			Version: 11,
		},
		{
			Id:      "app-two",
			Name:    "Bread",
			Version: 22,
		},
	}

	err = s.WriteApps(expectedApps)
	assert.Nil(t, err)

	actualApps, err := s.Apps()
	assert.Nil(t, err)

	assert.Equal(t, expectedApps, actualApps)

	expectedTokens := []model.Token{
		{
			AccountType: "fake",
			Name:        "fake",
			Digits:      6,
		},
	}

	err = s.WriteTokens(expectedTokens)
	assert.Nil(t, err)

	actualTokens, err := s.Tokens()
	assert.Nil(t, err)

	assert.Equal(t, expectedTokens, actualTokens)
}
