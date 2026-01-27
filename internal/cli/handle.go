package cli

import (
	"bufio"
	"slices"
	"strings"
)

func ReadLine(r any) (string, error) {
	reader, ok := r.(interface{ ReadString(byte) (string, error) })
	if !ok {
		br, ok := r.(interface{ ReadBytes(byte) ([]byte, error) })
		if !ok {
			scanner := bufio.NewScanner(r.(interface{ Read([]byte) (int, error) }))
			if scanner.Scan() {
				return scanner.Text(), nil
			}
			return "", scanner.Err()
		}
		line, err := br.ReadBytes('\n')
		return strings.TrimSuffix(string(line), "\n"), err
	}
	line, err := reader.ReadString('\n')
	return strings.TrimSuffix(line, "\n"), err
}

func ExtractHandleFromArgs(args []string, extraFilters ...string) (handle string, remaining []string) {
	handleSet := false
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			remaining = append(remaining, arg)
			continue
		}

		if slices.Contains(extraFilters, arg) {
			remaining = append(remaining, arg)
			continue
		}

		if !handleSet {
			handle = arg
			handleSet = true
		} else {
			remaining = append(remaining, arg)
		}
	}

	return handle, remaining
}
