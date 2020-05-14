package envparser

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"testing"

	. "github.com/onsi/gomega"
)

type testFS struct {
	returnVal io.ReadCloser
	returnErr error
}

func (tfs *testFS) Open(path string) (io.ReadCloser, error) {
	if tfs.returnErr != nil {
		return nil, errors.New("error in testFS")
	}

	return tfs.returnVal, nil
}

type badReadCloser struct{}

func (b *badReadCloser) Read(p []byte) (int, error) {
	return 0, errors.New("bad read")
}
func (b *badReadCloser) Close() error {
	return nil
}

func TestParseFile(t *testing.T) {
	tt := []struct {
		name    string
		envFile string
		plain   map[string]string
		secrets map[string]string
	}{
		{
			name: "only plain",
			envFile: `
				PLAIN=plaintext
				PLAIN_ZERO=subzero
			`,
			plain: map[string]string{
				"PLAIN":      "plaintext",
				"PLAIN_ZERO": "subzero",
			},
			secrets: map[string]string{},
		},
		{
			name: "comments with plain",
			envFile: `
				#comment
				PLAIN=plaintext
				#now he's
				PLAIN_ZERO=subzero
				#PLAIN=another
			`,
			plain: map[string]string{
				"PLAIN":      "plaintext",
				"PLAIN_ZERO": "subzero",
			},
			secrets: map[string]string{},
		},
		{
			name: "plain and secrets",
			envFile: `
				PLAIN=plaintext
				SECRET=!{keep it secret, keep it safe}
			`,
			plain: map[string]string{
				"PLAIN": "plaintext",
			},
			secrets: map[string]string{
				"SECRET": "keep it secret, keep it safe",
			},
		},
		{
			name: "plain secrets and comments",
			envFile: `
				#yoyo
				PLAIN=plaintext
				SECRET=!{keep it secret, keep it safe}
			`,
			plain: map[string]string{
				"PLAIN": "plaintext",
			},
			secrets: map[string]string{
				"SECRET": "keep it secret, keep it safe",
			},
		},
		{
			name: "only secrets and comments",
			envFile: `
				#yoyo
				SECRET=!{keep it secret, keep it safe}
			`,
			plain: map[string]string{},
			secrets: map[string]string{
				"SECRET": "keep it secret, keep it safe",
			},
		},
		{
			name: "only secrets",
			envFile: `
				SECRET_PASSWORD=!{is it the red or the white?}
			`,
			plain: map[string]string{},
			secrets: map[string]string{
				"SECRET_PASSWORD": "is it the red or the white?",
			},
		},
		{
			name: "= in value",
			envFile: `
				EQUAL=1+1=3
			`,
			plain: map[string]string{
				"EQUAL": "1+1=3",
			},
			secrets: map[string]string{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			retVal := ioutil.NopCloser(bytes.NewBufferString(tc.envFile))
			tfs := &testFS{returnVal: retVal}
			env, err := parseFile(tfs, "")
			g.Expect(err).To(BeNil())

			g.Expect(env).ToNot(BeNil())
			g.Expect(env.Plain).To(Equal(tc.plain))
			g.Expect(env.Secrets).To(Equal(tc.secrets))
		})
	}
}

func TestParseFileError(t *testing.T) {
	tt := []struct {
		name        string
		envFile     string
		returnErr   error
		expectedErr error
	}{
		{
			name:        "open file error",
			returnErr:   errors.New("file error"),
			expectedErr: errors.New("error opening file"),
		},
		{
			name: "no key value",
			envFile: `
				BAD_VALUE
			`,
			expectedErr: errors.New("invalid format"),
		},
		{
			name: "no key",
			envFile: `
				=no key
			`,
			expectedErr: errors.New("invalid format"),
		},
		{
			name:        "only kv seperator",
			envFile:     kvSeparator,
			expectedErr: errors.New("invalid format"),
		},
		{
			name: "empty secret",
			envFile: `
				SECRET=!{}
			`,
			expectedErr: errors.New("invalid format"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			tfs := &testFS{
				returnVal: ioutil.NopCloser(bytes.NewBufferString(tc.envFile)),
				returnErr: tc.returnErr,
			}

			env, err := parseFile(tfs, "")
			g.Expect(env).To(BeNil())

			g.Expect(err).ToNot(BeNil())
			g.Expect(err.Error()).To(ContainSubstring(tc.expectedErr.Error()))
		})
	}
}

func TestParseFileScannerError(t *testing.T) {
	g := NewGomegaWithT(t)

	tfs := &testFS{returnVal: &badReadCloser{}}
	env, err := parseFile(tfs, "")
	g.Expect(env).To(BeNil())

	g.Expect(err).ToNot(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("error parsing file"))
}
