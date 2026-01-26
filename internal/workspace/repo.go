package workspace

import "strings"

func ParseRepoFlag(repo string) (url, ref string) {
	repo = strings.TrimSpace(repo)

	if strings.HasPrefix(repo, "git@") {
		colonIdx := strings.Index(repo, ":")
		if colonIdx != -1 {
			atIdx := strings.LastIndex(repo[colonIdx:], "@")
			if atIdx != -1 {
				actualIdx := colonIdx + atIdx
				url = repo[:actualIdx]
				ref = repo[actualIdx+1:]
				return url, ref
			}
		}
		return repo, ""
	}

	atIdx := strings.LastIndex(repo, "@")
	if atIdx != -1 {
		url = repo[:atIdx]
		ref = repo[atIdx+1:]
	} else {
		url = repo
	}

	return url, ref
}
