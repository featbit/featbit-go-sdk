package fixtures

import (
	"github.com/featbit/featbit-go-sdk/v2/internal/util"
	"os"
	"path"
)

func LoadFBClientTestData() ([]byte, error) {
	// get root absolute path
	root, err := os.Getwd()
	if err != nil {
		return []byte(nil), err
	}
	return util.ReadFile(path.Join(root, "fixtures", "fbclient_test_data.json"))
}
