package libs

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"os"
)

const csvData = `"service", "name", "goal"
tsat, drift-gen, 0.99
`

func UploadSLOLookup(id, url, filename string) error {
	//filetype := "csv"
	//file, err := os.ReadFile(filename)

	client := resty.New()

	resp, err := client.R().SetBasicAuth(os.Getenv(EnvKeySumoAccessID), os.Getenv(EnvKeySumoAccessKey)).
		SetFile("file", filename).
		Post(url)

	fmt.Println(resp.Status(), resp.String())

	return err
}
