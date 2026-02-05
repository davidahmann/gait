package main

import "strings"

func reorderInterspersedFlags(arguments []string, valueFlags map[string]bool) []string {
	if len(arguments) == 0 {
		return arguments
	}

	flags := make([]string, 0, len(arguments))
	positionals := make([]string, 0, len(arguments))

	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		if argument == "--" {
			positionals = append(positionals, arguments[index+1:]...)
			break
		}
		if !isFlagToken(argument) {
			positionals = append(positionals, argument)
			continue
		}

		flags = append(flags, argument)
		if strings.Contains(argument, "=") {
			continue
		}
		if !flagRequiresValue(argument, valueFlags) {
			continue
		}
		if index+1 >= len(arguments) {
			continue
		}
		index++
		flags = append(flags, arguments[index])
	}

	return append(flags, positionals...)
}

func isFlagToken(argument string) bool {
	return len(argument) > 1 && strings.HasPrefix(argument, "-")
}

func flagRequiresValue(argument string, valueFlags map[string]bool) bool {
	if len(valueFlags) == 0 {
		return false
	}
	if required, ok := valueFlags[argument]; ok {
		return required
	}

	name := strings.TrimLeft(argument, "-")
	required, ok := valueFlags[name]
	return ok && required
}
