package crypto

import (
	"strings"
)

// MaskConnectionURI masks sensitive user:password credentials inside standard connection URIs.
// e.g. postgresql://dbuser:MyPassword@localhost:5432/db -> postgresql://dbuser:********@localhost:5432/db
func MaskConnectionURI(uri string) string {
	if !strings.Contains(uri, "@") || !strings.Contains(uri, "://") {
		return uri
	}

	protoSplit := strings.SplitN(uri, "://", 2)
	if len(protoSplit) < 2 {
		return uri
	}
	protocol := protoSplit[0]
	rest := protoSplit[1]

	authSplit := strings.SplitN(rest, "@", 2)
	if len(authSplit) < 2 {
		return uri
	}
	auth := authSplit[0]
	hostPart := authSplit[1]

	if strings.Contains(auth, ":") {
		userSplit := strings.SplitN(auth, ":", 2)
		user := userSplit[0]
		return protocol + "://" + user + ":********@" + hostPart
	}

	return uri
}
