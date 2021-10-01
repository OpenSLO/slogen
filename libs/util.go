package libs

import (
	"errors"
	"os"
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
