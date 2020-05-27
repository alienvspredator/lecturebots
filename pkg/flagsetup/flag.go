package flagsetup

import (
	"flag"
	"fmt"
	"strings"
)

// CheckRequired check is the required flags set.
func CheckRequired(requiredFlags []string) error {
	reqFlMap := makeRequiredFlagMap(requiredFlags)

	flag.Visit(func(fl *flag.Flag) {
		reqFlMap[fl.Name] = true
	})

	notSetup := make([]string, 0)
	for fl, ok := range reqFlMap {
		if !ok {
			notSetup = append(notSetup, fl)
		}
	}

	if len(notSetup) > 0 {
		return fmt.Errorf(
			"Expected required flags `%s`. Check -help",
			strings.Join(notSetup, ", "),
		)
	}

	return nil
}

func makeRequiredFlagMap(req []string) map[string]bool {
	m := make(map[string]bool)
	for _, flName := range req {
		m[flName] = false
	}

	return m
}
