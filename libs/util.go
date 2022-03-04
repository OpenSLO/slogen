package libs

import (
	"errors"
	"github.com/mitchellh/hashstructure/v2"
	"os"
	"sort"
	"strings"
)

func DeDupe(strSlice []string) []string {
	allKeys := make(map[string]bool)
	var list []string
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func EnsureDir(dirName string, clean bool) error {

	if clean {
		yes, _ := dirExists(dirName)

		var err error
		if yes {
			err = os.RemoveAll(dirName)
		}

		if err != nil {
			return err
		}
	}

	err := os.Mkdir(dirName, 0755)
	if err == nil {
		return nil
	}
	if os.IsExist(err) {
		// check that the existing path is a directory
		info, err := os.Stat(dirName)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return errors.New("path exists but is not a directory")
		}

		return nil
	}
	return err
}

func GiveKeys(m map[string]bool) []string {
	keys := []string{}

	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// exists returns whether the given file or directory exists
func dirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func giveMapKeys(m map[string]string) []string {
	keys := make([]string, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	return keys
}

func giveFieldsGroupByStr(m map[string]string) string {
	k := giveMapKeys(m)

	return strings.Join(k, ",")
}

func GiveStructCompare(a, b interface{}) bool {
	aHash, err := hashstructure.Hash(a, hashstructure.FormatV2, nil)
	if err != nil {
		log.Fatal(err)
	}
	bHash, err := hashstructure.Hash(b, hashstructure.FormatV2, nil)

	if err != nil {
		log.Fatal(err)
	}
	return aHash == bHash
}
