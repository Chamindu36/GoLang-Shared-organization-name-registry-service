package config

import (
	"bytes"
	"encoding"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Config struct {
	viper *viper.Viper
}

type validator interface {
	Validate() error
}

func MustReadFile(filename string) *Config {
	cfg, err := ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return cfg
}

func ReadFile(filename string) (*Config, error) {
	f, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return nil, err
	}
	return ReadConfig(f)
}

func ReadFileAsText(filename string) (string, error) {
	f, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return "", err
	}
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func ReadConfig(in io.Reader) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	r, err := replaceEnv(in)
	if err != nil {
		return nil, err
	}
	if err := v.ReadConfig(r); err != nil {
		return nil, err
	}
	return &Config{
		viper: v,
	}, nil
}

func replaceEnv(in io.Reader) (io.Reader, error) {
	rawCfgBuf := &strings.Builder{}
	_, err := io.Copy(rawCfgBuf, in)
	if err != nil {
		return nil, err
	}
	var missingEnvs []string
	replacedConfig := os.Expand(rawCfgBuf.String(), func(s string) string {
		v, ok := os.LookupEnv(s)
		if !ok {
			missingEnvs = append(missingEnvs, s)
			return ""
		}
		return v
	})

	if len(missingEnvs) > 0 {
		return nil, fmt.Errorf("missing env(s): %s", strings.Join(missingEnvs, ","))
	}

	return bytes.NewBuffer([]byte(replacedConfig)), nil
}

func (c *Config) IsSet(key string) bool {
	return c.viper.IsSet(key)
}

func (c *Config) Unmarshal(key string, out interface{}) error {
	err := c.viper.UnmarshalKey(key, out, func(c *mapstructure.DecoderConfig) {
		c.TagName = "yaml"
		c.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			StringToFunctionHookFunc(),
			StringToUrlHookFunc(),
		)
	})
	if err != nil {
		return fmt.Errorf("cannot unmarshal config for key %q: %v", key, err)
	}
	if v, ok := out.(validator); ok {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("invalid configuration for key %q: %v", key, err)
		}
	}
	return nil
}

func (c *Config) MustUnmarshal(key string, out interface{}) {
	err := c.Unmarshal(key, out)
	if err != nil {
		panic(err)
	}
}

func StringToFunctionHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || !(t.Kind() == reflect.Func || t.Kind() == reflect.Struct) {
			return data, nil
		}
		v := reflect.New(t)
		u, ok := v.Interface().(encoding.TextUnmarshaler)
		if ok {
			err := u.UnmarshalText([]byte(data.(string)))
			if err != nil {
				return nil, err
			}
			return v.Elem().Interface(), nil
		}
		return data, nil
	}
}

func StringToUrlHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		if t != reflect.TypeOf(url.URL{}) {
			return data, nil
		}

		return url.Parse(data.(string))
	}
}
