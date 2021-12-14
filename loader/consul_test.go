package loader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromConsul(t *testing.T) {
	type inner struct {
		Slice []string `yaml:"slice"`
	}
	type res struct {
		UntaggedStr     string
		CamelCaseStr    string `cfg:"camelCaseStr"`
		CamelCaseInt    int64  `cfg:"camelCaseInt"`
		CamelCaseStruct inner  `cfg:"camelCaseStruct"`
		SnakeCaseInt    int64  `cfg:"snake_case_int"`
		SnakeCaseStruct inner  `cfg:"snake_case_struct"`
	}

	tests := []struct {
		name       string
		consulConf ConsulMock
		to         res
		result     res
		err        string
	}{
		{
			name: "test-json",
			consulConf: ConsulMock{kv: map[string][]byte{
				"test-json": []byte(`{untaggedStr: 'untag value', camelCaseStr: 'camel case value', camelCaseInt: 64, camelCaseStruct: {slice: [one, two]}, snake_case_int: 55, snake_case_struct: {slice: [one, two, three four]}}`),
			}},
			result: res{
				UntaggedStr:     "untag value",
				CamelCaseStr:    "camel case value",
				CamelCaseInt:    64,
				CamelCaseStruct: inner{[]string{"one", "two"}},
				SnakeCaseInt:    55,
				SnakeCaseStruct: inner{[]string{"one", "two", "three four"}},
			},
		},
		{
			name: "test-yaml",
			consulConf: ConsulMock{kv: map[string][]byte{
				"test-yaml": []byte(`
untaggedStr: test
camelCaseStr: 'camel case value'
camelCaseInt: 64
camelCaseStruct:
  slice:
  - one
  - two
snake_case_int: 55
snake_case_struct:
  slice:
  - one
  - two
  - three four`),
			}},
			result: res{
				UntaggedStr:     "test",
				CamelCaseStr:    "camel case value",
				CamelCaseInt:    64,
				CamelCaseStruct: inner{[]string{"one", "two"}},
				SnakeCaseInt:    55,
				SnakeCaseStruct: inner{[]string{"one", "two", "three four"}},
			},
		},
		{
			name:       "no-key",
			consulConf: ConsulMock{kv: map[string][]byte{}},
		},
		{
			name:       "error",
			consulConf: ConsulMock{err: errors.New("test error")},
			err:        "test error",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Consul{Client: NewConsulMock(test.consulConf)}.Load(test.name, &test.to)

			if test.err == "" {
				assert.NoError(t, err)
			} else {
				// Errors from Consul client RoundTripper always wrapped with url.Error
				assert.EqualError(t, errors.Unwrap(err), test.err)
			}

			assert.Equal(t, test.result, test.to)
		})
	}
}

func TestNewConsuler_WrongAddr(t *testing.T) {
	c, err := NewConsul("locall:8787")

	assert.Nil(t, err)
	assert.NotNil(t, c)
}

func TestConsul_DynamicValue(t *testing.T) {
	t.Skip("currently DynamicValue is experimental")
	// Start with 5 so we will be able to output some same-value and same-index variables.
	var consulCalls = 5
	var configPath = path.Join("app", "field")

	zerolog.TimeFieldFormat = time.RFC3339Nano
	consuler := ConsulMock{kvFunc: func(keyPath string) (*api.KVPair, *api.QueryMeta, bool) {
		require.True(t, strings.HasPrefix(keyPath, configPath), "requested config path")

		consulCalls++

		if consulCalls > 6 {
			// Simulate waiting for new value.
			// Consul returns in two cases: when value is updated or on timeout.
			time.Sleep(201 * time.Millisecond)
		}

		return &api.KVPair{Key: keyPath, Value: []byte(strconv.Itoa(consulCalls / 5))},
			&api.QueryMeta{LastIndex: uint64(consulCalls / 5)},
			true
	}}

	consul := Consul{Client: NewConsulMock(consuler)}
	seenVals := map[string]int{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := consul.DynamicValue(ctx, DynamicConfig{
		AppName:   "app",
		FieldName: "field",
		Runner: func(value []byte) error {
			seenVals[string(value)]++

			return nil
		},
	})

	assert.Equal(t, 6, consulCalls-5)
	assert.Equal(t, map[string]int{"1": 1, "2": 1}, seenVals)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func NewConsulMock(mockConfig ConsulMock) *api.Client {
	cl, _ := api.NewClient(&api.Config{
		HttpClient: &http.Client{
			Transport: mockConfig,
		},
	})

	return cl
}

type ConsulMock struct {
	kvFunc func(keyPath string) (*api.KVPair, *api.QueryMeta, bool)
	kv     map[string][]byte
	err    error
}

func (m ConsulMock) RoundTrip(request *http.Request) (*http.Response, error) {
	reqURI := request.URL.RequestURI()

	switch {
	case strings.HasPrefix(reqURI, "/v1/kv/"):
		key := strings.TrimPrefix(reqURI, path.Join("/v1/kv", ConsulConfigPathPrefix)+"/")

		kvResp, meta, err := m.Get(key, nil)

		bts, _ := json.Marshal(api.KVPairs{kvResp})

		httpResp := http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(bts)),
		}

		if kvResp == nil {
			httpResp.StatusCode = http.StatusNotFound
		}

		httpResp.Header = generateMetaHeader(meta)

		return &httpResp, err
	}

	return nil, fmt.Errorf("%s %s", request.Method, request.URL.RequestURI())
}

func (m ConsulMock) Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error) {
	if m.err != nil {
		return nil, nil, m.err
	}

	var data = &api.KVPair{
		Key: key,
	}
	var meta *api.QueryMeta

	var ok bool

	data.Value, ok = m.kv[key]
	if !ok {
		if m.kvFunc == nil {
			return nil, nil, nil
		}

		data, meta, ok = m.kvFunc(key)
	}

	return data, meta, nil
}

func generateMetaHeader(meta *api.QueryMeta) http.Header {
	var h = http.Header{}

	if meta == nil {
		return h
	}

	h.Set("X-Consul-Index", strconv.FormatUint(meta.LastIndex, 10))

	return h
}
