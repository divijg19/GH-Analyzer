package main

import (
	"flag"
	"fmt"
)

func resolveLimitFlag(fs *flag.FlagSet, longValue, shortValue int) (int, error) {
	longSet := false
	shortSet := false

	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "limit":
			longSet = true
		case "k":
			shortSet = true
		}
	})

	switch {
	case longSet && shortSet && longValue != shortValue:
		return 0, fmt.Errorf("conflicting --limit (%d) and -k (%d)", longValue, shortValue)
	case shortSet:
		return shortValue, nil
	default:
		return longValue, nil
	}
}
