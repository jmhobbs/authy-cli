package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/jmhobbs/authy-cli/model"
	"golang.org/x/crypto/argon2"
)

var ErrNotFound = errors.New("Not Found")

type filename string

const (
	configFile filename = "config"
	appsFile   filename = "apps"
	tokensFile filename = "tokens"
)

type PassphrasePromptFunc func() (string, error)

type Store struct {
	prompt      PassphrasePromptFunc
	_passphrase *string
	root        string
}

func New(root string, prompt PassphrasePromptFunc) (*Store, error) {
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, fmt.Errorf("unable to create storage directory: %w", err)
	}

	return &Store{
		prompt:      prompt,
		_passphrase: nil,
		root:        root,
	}, nil
}

func (s *Store) fetchPassphraseWithCache() (string, error) {
	if s._passphrase != nil {
		return *s._passphrase, nil
	}

	passphrase, err := s.prompt()
	if err != nil {
		return "", err
	}

	s._passphrase = &passphrase
	return passphrase, nil
}

func (s *Store) loadAndDecrypt(file filename) ([]byte, error) {
	f, err := os.Open(path.Join(s.root, string(file)))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	passphrase, err := s.fetchPassphraseWithCache()
	if err != nil {
		return nil, err
	}

	// read the passphrase salt off the front
	var salt [16]byte
	_, err = f.Read(salt[:])
	if err != nil {
		return nil, fmt.Errorf("unable to read salt: %w", err)
	}

	// generate the key from passphrase + salt
	var key [32]byte
	copy(key[:], argon2.IDKey(
		[]byte(passphrase),
		salt[:],
		1,         // time
		2048*1024, // memory (2GB)
		4,         // threads
		32,        // key length
	))

	// decrypt
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// read the nonce
	var nonce []byte = make([]byte, gcm.NonceSize())
	_, err = f.Read(nonce)
	if err != nil {
		return nil, fmt.Errorf("unable to read nonce: %w", err)
	}

	// read the rest as the ciphertext
	ciphertext, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, ciphertext, nil)
}

func (s *Store) encryptAndStore(plaintext []byte, file filename) error {
	passphrase, err := s.fetchPassphraseWithCache()
	if err != nil {
		return err
	}

	// generate a salt
	var salt [16]byte
	if _, err := io.ReadFull(rand.Reader, salt[:]); err != nil {
		return err
	}

	// generate the key from passphrase + salt
	var key [32]byte
	copy(key[:], argon2.IDKey(
		[]byte(passphrase),
		salt[:],
		1,         // time
		2048*1024, // memory (2GB)
		4,         // threads
		32,        // key length
	))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// generate a nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	// encrypt
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// write the file
	f, err := os.OpenFile(path.Join(s.root, string(file)), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// write the salt
	_, err = f.Write(salt[:])
	if err != nil {
		return err
	}

	// write the nonce
	_, err = f.Write(nonce)
	if err != nil {
		return err
	}

	// write the ciphertext
	_, err = f.Write(ciphertext)
	return err
}

func (s *Store) Config() (*Config, error) {
	plaintext, err := s.loadAndDecrypt(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, ErrNotFound
		}
	}

	var config Config
	err = json.Unmarshal(plaintext, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (s *Store) WriteConfig(config Config) error {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return s.encryptAndStore(configBytes, configFile)
}

func (s *Store) Apps() ([]model.App, error) {
	plaintext, err := s.loadAndDecrypt(appsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.App{}, ErrNotFound
		}
	}

	var apps []model.App
	err = json.Unmarshal(plaintext, &apps)
	if err != nil {
		return nil, err
	}

	return apps, nil
}

func (s *Store) WriteApps(apps []model.App) error {
	appsBytes, err := json.Marshal(apps)
	if err != nil {
		return err
	}
	return s.encryptAndStore(appsBytes, appsFile)
}

func (s *Store) Tokens() ([]model.Token, error) {
	plaintext, err := s.loadAndDecrypt(tokensFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.Token{}, ErrNotFound
		}
	}

	var tokens []model.Token
	err = json.Unmarshal(plaintext, &tokens)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *Store) WriteTokens(tokens []model.Token) error {
	tokensBytes, err := json.Marshal(tokens)
	if err != nil {
		return err
	}
	return s.encryptAndStore(tokensBytes, tokensFile)
}

func (s *Store) App(search string) (model.App, error) {
	apps, err := s.Apps()
	if err != nil {
		return model.App{}, err
	}
	for _, app := range apps {
		if app.Name == search || app.Id == search {
			return app, nil
		}
	}
	return model.App{}, ErrNotFound
}

func (s *Store) Token(search string) (model.Token, error) {
	tokens, err := s.Tokens()
	if err != nil {
		return model.Token{}, err
	}
	for _, token := range tokens {
		if token.UniqueId == search || token.Name == search {
			return token, nil
		}
	}
	return model.Token{}, ErrNotFound
}
