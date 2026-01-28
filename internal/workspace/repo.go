package workspace

import (
	"strconv"
	"strings"
)

func ParseRepoFlag(repo string) (url, ref string, depth int) {
	repo = strings.TrimSpace(repo)

	baseRepo := repo
	depth = 0

	doubleColonIdx := strings.LastIndex(repo, "::")
	if doubleColonIdx != -1 {
		depthPart := repo[doubleColonIdx+2:]
		if d, err := strconv.Atoi(depthPart); err == nil {
			depth = d
			baseRepo = repo[:doubleColonIdx]
		}
	}

	if strings.HasPrefix(baseRepo, "git@") {
		atIdx := strings.Index(baseRepo, ":")
		if atIdx != -1 {
			afterAt := baseRepo[atIdx+1:]
			gitAtIdx := strings.Index(afterAt, "@")
			if gitAtIdx != -1 {
				url = baseRepo[:atIdx+1+gitAtIdx]
				ref = afterAt[gitAtIdx+1:]
				return url, ref, depth
			}
		}
		url = baseRepo
		ref = ""
		return url, ref, depth
	}

	atIdx := strings.LastIndex(baseRepo, "@")
	if atIdx != -1 {
		url = baseRepo[:atIdx]
		ref = baseRepo[atIdx+1:]
	} else {
		url = baseRepo
		ref = ""
	}

	return url, ref, depth
}
