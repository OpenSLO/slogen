package libs

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// ValidateDir returns a non-nil error if invalid along with file name

type Results struct {
	file string
	err  error
}

func ParseDir(path string, ignoreErrors bool) (map[string]*SLO, error) {

	slos := make(map[string]*SLO)

	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			if ignoreErrors {
				return nil
			}
			return err
		}

		name := info.Name()
		if info.IsDir() {
			WarnInfo("\nentering dir : %s\n", name)
			return nil
		}
		if !isYaml(name) {
			WarnInfo("\nignoring non-yaml file : ")
			WarnUInfo("%s\n", path)
			return nil
		}

		slo, err := ParseFile(path)

		if err != nil {
			if ignoreErrors {
				return nil
			}
			return err
		}

		slos[path] = slo
		return nil
	})

	return slos, err
}

func ParseFile(path string) (*SLO, error) {

	slo, err := Parse(path)

	if err != nil {
		BadInfo("\nError : %s\n", err)
		BadResult("\nInvalid Config : ")
		BadUResult("%s\n", path)
		return nil, err
	}

	GoodResult("\nvalid config : ")
	GoodUResult("%s\n", path)
	return slo, err
}

func isYaml(name string) bool {
	return strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")
}
