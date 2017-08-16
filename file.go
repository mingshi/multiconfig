package multiconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/arstd/log"
	yaml "gopkg.in/yaml.v2"
)

var (
	// ErrSourceNotSet states that neither the path or the reader is set on the loader
	ErrSourceNotSet = errors.New("config path or reader is not set")

	// ErrFileNotFound states that given file is not exists
	ErrFileNotFound = errors.New("config file not found")
)

// TOMLLoader satisifies the loader interface. It loads the configuration from
// the given toml file or Reader.
type TOMLLoader struct {
	Path   string
	Reader io.Reader
}

// Load loads the source into the config defined by struct s
// Defaults to using the Reader if provided, otherwise tries to read from the
// file
func (t *TOMLLoader) Load(s interface{}) error {
	var r io.Reader

	if t.Reader != nil {
		r = t.Reader
	} else if t.Path != "" {
		file, err := getConfig(t.Path)
		if err != nil {
			return err
		}
		defer file.Close()
		r = file
	} else {
		return ErrSourceNotSet
	}

	if _, err := toml.DecodeReader(r, s); err != nil {
		return err
	}

	return nil
}

// JSONLoader satisifies the loader interface. It loads the configuration from
// the given json file or Reader.
type JSONLoader struct {
	Path   string
	Reader io.Reader
}

// Load loads the source into the config defined by struct s.
// Defaults to using the Reader if provided, otherwise tries to read from the
// file
func (j *JSONLoader) Load(s interface{}) error {
	var r io.Reader
	if j.Reader != nil {
		r = j.Reader
	} else if j.Path != "" {
		file, err := getConfig(j.Path)
		if err != nil {
			return err
		}
		defer file.Close()
		r = file
	} else {
		return ErrSourceNotSet
	}

	return json.NewDecoder(r).Decode(s)
}

// YAMLLoader satisifies the loader interface. It loads the configuration from
// the given yaml file.
type YAMLLoader struct {
	Path   string
	Reader io.Reader
}

// Load loads the source into the config defined by struct s.
// Defaults to using the Reader if provided, otherwise tries to read from the
// file
func (y *YAMLLoader) Load(s interface{}) error {
	var r io.Reader

	if y.Reader != nil {
		r = y.Reader
	} else if y.Path != "" {
		file, err := getConfig(y.Path)
		if err != nil {
			return err
		}
		defer file.Close()
		r = file
	} else {
		return ErrSourceNotSet
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, s)
}

func getConfig(path string) (f *os.File, err error) {
	if filepath.IsAbs(path) {
		f, err = os.Open(path)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file %s not exist", path)
		}
		log.Infof("read conf file %s", path)
		return f, err
	}

	var pwd string
	pwd, err = os.Getwd()
	if err != nil {
		return nil, err
	}

	for {
		if len(pwd) <= 1 {
			return nil, fmt.Errorf("file %s not exist", path)
		}

		abs := filepath.Join(pwd, path)
		f, err = os.Open(abs)
		if err == nil {
			log.Infof("read conf file %s", abs)
			return f, nil
		}

		if !os.IsNotExist(err) {
			return nil, err
		}

		i := strings.LastIndexByte(pwd, filepath.Separator)
		if i == -1 {
			return nil, fmt.Errorf("file %s not exist", path)
		}
		pwd = pwd[:i]
	}
}
