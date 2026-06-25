package routes

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
)

var (
	convertedSpec []byte
	convertOnce   sync.Once
	convertErr    error
)

func getOpenAPI3Spec() ([]byte, error) {
	convertOnce.Do(func() {
		// Read the original Swagger 2.0 file
		data, err := os.ReadFile("docs/swagger.json")
		if err != nil {
			convertErr = err
			return
		}

		// Parse Swagger 2.0
		var swagger2 openapi2.T
		if err = json.Unmarshal(data, &swagger2); err != nil {
			convertErr = err
			return
		}

		// Convert to OpenAPI 3.0
		doc3, err := openapi2conv.ToV3(&swagger2)
		if err != nil {
			convertErr = err
			return
		}

		// Marshal back to JSON
		convertedSpec, convertErr = json.MarshalIndent(doc3, "", "  ")
	})
	return convertedSpec, convertErr
}
