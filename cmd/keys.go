/*
 * Copyright 2015 Casey Marshall
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/crypto/nacl/secretbox"
	"gopkg.in/macaroon-bakery.v1/bakery"
)

var defaultHome string

func init() {
	var userHomeDir string
	if u, err := user.Current(); err == nil {
		userHomeDir = u.HomeDir
	}
	if userHomeDir == "" {
		userHomeDir = os.Getenv("HOME")
	}
	if userHomeDir != "" {
		defaultHome = filepath.Join(userHomeDir, ".oo")
	}
}

type keyPair struct {
	*bakery.KeyPair
}

func newKeyPair() *keyPair {
	return &keyPair{&bakery.KeyPair{}}
}

func (kp *keyPair) load(keyPath string) error {
	f, err := os.Open(keyPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&kp.KeyPair)
}

func (kp *keyPair) save(keyPath string) error {
	f, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(kp.KeyPair)
}

type keyManager struct {
	Context
}

func (m keyManager) homeDir() (string, error) {
	home := filepath.FromSlash(m.Context.String("home"))
	if home == "" {
		return "", fmt.Errorf("could not determine OO_HOME, --home is required")
	}
	return home, nil
}

func (m keyManager) keyPath() (string, error) {
	home, err := m.homeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "key"), nil
}

func (m keyManager) keyPair() (*keyPair, error) {
	keyPath, err := m.keyPath()
	if err != nil {
		return nil, err
	}
	kp := newKeyPair()
	if err = kp.load(keyPath); err == nil {
		return kp, nil
	} else if os.IsNotExist(err) {
		bakeryKeyPair, err := bakery.GenerateKey()
		if err != nil {
			return nil, fmt.Errorf("failed to create new key pair: %v", err)
		}
		kp.KeyPair = bakeryKeyPair
		err = kp.save(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to save new key pair: %v", err)
		}
		return kp, nil
	}
	return nil, err
}

type envelope struct {
	nonce  *[24]byte
	key    *[32]byte
	sha384 [sha512.Size384]byte
}

func newEnvelope() *envelope {
	return &envelope{nonce: new([24]byte), key: new([32]byte)}
}

func generateEnvelope() (*envelope, error) {
	nonce := new([24]byte)
	_, err := rand.Reader.Read(nonce[:])
	if err != nil {
		return nil, err
	}
	key := new([32]byte)
	_, err = rand.Reader.Read(key[:])
	if err != nil {
		return nil, err
	}
	return &envelope{nonce: nonce, key: key}, nil
}

func (e *envelope) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Nonce, Key, SHA384 []byte
	}{e.nonce[:], e.key[:], e.sha384[:]})
}

func (e *envelope) UnmarshalJSON(buf []byte) error {
	var st struct {
		Nonce, Key, SHA384 []byte
	}
	err := json.Unmarshal(buf, &st)
	if err != nil {
		return err
	}

	if len(st.Nonce) != 24 {
		return fmt.Errorf("invalid nonce length %d", len(st.Nonce))
	}
	copy(e.nonce[:], st.Nonce)

	if len(st.Key) != 32 {
		return fmt.Errorf("invalid key length %d", len(st.Key))
	}
	copy(e.key[:], st.Key)

	if len(st.SHA384) != sha512.Size384 {
		return fmt.Errorf("invalid digest length %d", len(st.SHA384))
	}
	copy(e.sha384[:], st.SHA384)

	return nil
}

func encrypt(r io.ReadCloser) (*envelope, io.ReadCloser, error) {
	var contents bytes.Buffer
	_, err := io.Copy(&contents, r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read content: %v", err)
	}
	err = r.Close()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to close input: %v", err)
	}

	digest := sha512.Sum384(contents.Bytes())
	env, err := generateEnvelope()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create envelope: %v", err)
	}
	env.sha384 = digest
	out := secretbox.Seal(nil, contents.Bytes(), env.nonce, env.key)
	// TODO: zeroize `contents`
	return env, ioutil.NopCloser(bytes.NewBuffer(out)), nil
}

func (env *envelope) decrypt(r io.Reader) (io.Reader, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	out, ok := secretbox.Open(nil, buf, env.nonce, env.key)
	if !ok {
		return nil, fmt.Errorf("decryption failed")
	}
	return bytes.NewBuffer(out), nil
}
