package azureservicebus

import (
	"fmt"
	"testing"
)

func TestParseValidConnectionString(t *testing.T) {
	namespace := "test"

	cnx, err := ParseConnectionString(fmt.Sprintf("Endpoint=sb://%s.servicebus.windows.net/;SharedAccessKeyName=TestSharedAccessKey;SharedAccessKey=TestSharedAccessKey", namespace))
	if err != nil {
		t.Errorf("Valid connectionstring could not parsed.")
	}

	if cnx == nil {
		t.Errorf("Valid connectionstring was parsed to nil instance.")
	}

	if cnx != nil && cnx.namespace != namespace {
		t.Errorf("Namespace doesn't match after parsing.")
	}
}

func TestParseInvalidConnectionStringMissingEndpoint(t *testing.T) {
	cnx, err := ParseConnectionString("SharedAccessKeyName=TestSharedAccessKey;SharedAccessKey=TestSharedAccessKey")
	if err == nil {
		t.Errorf("Invalid connectionstring was happily parsed. Connectionstring missing `Endpoint` value.")
	}

	if cnx != nil && cnx.namespace != "" {
		t.Errorf("Namespace exists after parsing empty `Endpoint`.")
	}
}

func TestParseInvalidConnectionStringMissingKeyName(t *testing.T) {
	cnx, err := ParseConnectionString("Endpoint=sb://test.servicebus.windows.net/;SharedAccessKey=TestSharedAccessKey")
	if err == nil {
		t.Errorf("Invalid connectionstring was happily parsed. Connectionstring missing `SharedAccessKeyName` parameter.")
	}

	if cnx != nil && cnx.keyName != "" {
		t.Errorf("Key name exists after parsing empty `SharedAccessKeyName`.")
	}
}

func TestParseInvalidConnectionStringMissingKeyValue(t *testing.T) {
	cnx, err := ParseConnectionString("Endpoint=sb://test.servicebus.windows.net/;SharedAccessKeyName=TestSharedAccessKey")
	if err == nil {
		t.Errorf("Invalid connectionstring was happily parsed. Connectionstring missing `SharedAccessKey` parameter.")
	}

	if cnx != nil && cnx.accessKey != "" {
		t.Errorf("Key value exists after parsing empty `SharedAccessKey`.")
	}
}

func TestParseInvalidConnectionStringInvalidEndpoint(t *testing.T) {
	cnx, err := ParseConnectionString("Endpoint=no-valid-uri;SharedAccessKeyName=TestSharedAccessKey;SharedAccessKey=TestSharedAccessKey")
	if err == nil {
		t.Errorf("Invalid connectionstring was happily parsed. Connectionstring contains invalid `Endpoint` value.")
	}

	if cnx != nil && cnx.namespace != "" {
		t.Errorf("Namespace exists after parsing invalid `Endpoint`.")
	}
}
