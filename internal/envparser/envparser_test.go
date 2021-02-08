package envparser

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"testing"

	"gotest.tools/v3/assert"
)

type testFS struct {
	returnVal io.ReadCloser
	returnErr error
}

func (tfs *testFS) Open(path string) (io.ReadCloser, error) {
	if tfs.returnErr != nil {
		return nil, tfs.returnErr
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
		{
			name: "whitespace in value",
			envFile: `
				CAKE=   is a lie
			`,
			plain: map[string]string{
				"CAKE": "   is a lie",
			},
			secrets: map[string]string{},
		},
		{
			name: "multiline variables",
			envFile: `
				PLAIN=plaintext
				JWT_KEY="-----BEGIN PUBLIC KEY-----
MIGbMBAGByqGSM49AgEGBSuBBAAjA4GGAAQAC6vH7IGAp8pdUt92yiDGKt9mAwN3
TRT3ZSQfIk/68btXJRMBz1yqTYdjjmruG/H9uGq4N4g++djUb3k18Ep0MbsB6g+7
Dpig7Mu3Nqf3ywLsiXf1EiffYsrkUouWsjTnIYf800jl/JHHB0Gkn24td8aahE8v
5fK56W2mjskKKKCZnMc=
-----END PUBLIC KEY-----"
				PLAIN2=plaintext2
			`,
			plain: map[string]string{
				"PLAIN":  "plaintext",
				"PLAIN2": "plaintext2",
				"JWT_KEY": `-----BEGIN PUBLIC KEY-----
MIGbMBAGByqGSM49AgEGBSuBBAAjA4GGAAQAC6vH7IGAp8pdUt92yiDGKt9mAwN3
TRT3ZSQfIk/68btXJRMBz1yqTYdjjmruG/H9uGq4N4g++djUb3k18Ep0MbsB6g+7
Dpig7Mu3Nqf3ywLsiXf1EiffYsrkUouWsjTnIYf800jl/JHHB0Gkn24td8aahE8v
5fK56W2mjskKKKCZnMc=
-----END PUBLIC KEY-----`,
			},
			secrets: map[string]string{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			retVal := ioutil.NopCloser(bytes.NewBufferString(tc.envFile))
			tfs := &testFS{returnVal: retVal}

			env, err := parseFile(tfs, "")
			assert.NilError(t, err)
			assert.Assert(t, env != nil)
			assert.DeepEqual(t, env.Plain, tc.plain)     //nolint:staticcheck
			assert.DeepEqual(t, env.Secrets, tc.secrets) //nolint:staticcheck
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
			name:        "only kv separator",
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
			tfs := &testFS{
				returnVal: ioutil.NopCloser(bytes.NewBufferString(tc.envFile)),
				returnErr: tc.returnErr,
			}

			env, err := parseFile(tfs, "")
			assert.Assert(t, env == nil)
			assert.Assert(t, err != nil)
			assert.ErrorContains(t, err, tc.expectedErr.Error())
		})
	}
}

func TestParseFileScannerError(t *testing.T) {
	tfs := &testFS{returnVal: &badReadCloser{}}

	env, err := parseFile(tfs, "")
	assert.Assert(t, env == nil)
	assert.Assert(t, err != nil)
	assert.ErrorContains(t, err, "error parsing file")
}
