package ReceiveSyslog

import (
	"io/ioutil"
	"encoding/json"
	"testing"

	"github.com/TIBCOSoftware/flogo-lib/core/trigger"
)

func getJsonMetadata() string {
	jsonMetadataBytes, err := ioutil.ReadFile("trigger.json")
	if err != nil {
		panic("No Json Metadata found for trigger.json path")
	}
	return string(jsonMetadataBytes)
}

const testConfig string = `{
  "id": "mytrigger",
  "settings": {
    "port": "10514"
  },
  "handlers": [
    {
      "settings": {
        "regex": ".*"
      },
      "action" {
	     "id": "test_action"
      }
    }
  ]
}`

func TestCreate(t *testing.T) {

	// New factory
	md := trigger.NewMetadata(getJsonMetadata())
	f := NewFactory(md)

	if f == nil {
		t.Fail()
	}

	// New Trigger
	config := trigger.Config{}
	json.Unmarshal([]byte(testConfig), config)
	trg := f.New(&config)

	if trg == nil {
		t.Fail()
	}
}
