package fixtures

import (
	"github.com/featbit/featbit-go-sdk/internal/util"
	"github.com/featbit/featbit-go-sdk/internal/util/log"
	"os"
	"path"
)

func LoadFBClientTestData() []byte {
	// get root absolute path
	root, err := os.Getwd()
	if err != nil {
		log.LogError("FB GO SDK: error loading file - %v", err)
	}
	return util.ReadFile(path.Join(root, "fixtures", "fbclient_test_data.json"))
}
