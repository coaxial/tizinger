// Package credentials abstracts away access to the data contained withing the credentials.yml file.
package credentials

import (
	"io/ioutil"
	"sync"

	"github.com/coaxial/tizinger/utils/logger"
	"gopkg.in/yaml.v3"
)

// credentialsYAML represents the credentials.yml file's YAML structure.
type credentialsYAML struct {
	Tidal []TidalAccount `yaml:"tidal,omitempty"`
}

// TidalAccount represents credentials for the Tidal streaming service.
type TidalAccount struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// credentials holds the unmarshalled credentials.yml file contents.
var accounts credentialsYAML

// credentialsFile is the path to the credentials.yml file.
var credentialsFile = "credentials.yaml"

// once ensures the credentials file is loaded and parsed from disk
// only once, to avoid reading and parsing it every time credentials are
// requested.
var once sync.Once

// loadConfig reads and unmarshalls the credentials file.
func loadConfig() {
	logger.Trace.Printf("reading credentials from %q", credentialsFile)
	content, err := ioutil.ReadFile(credentialsFile)
	if err != nil {
		logger.Error.Fatalf("could not read %q: %v", credentialsFile, err)
		return
	}

	err = yaml.Unmarshal(content, &accounts)
	if err != nil {
		logger.Error.Fatalf("could not parse %q: %v", credentialsFile, err)
		return
	}
}

// Tidal exposes the tidal accounts credentials set in credentials.yaml.
func Tidal() (tc []TidalAccount, err error) {
	once.Do(loadConfig)
	return accounts.Tidal, err
}
