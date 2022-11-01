package gcp

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"flightcrew.io/cli/internal/debug"
)

const (
	prodHostName      = "api.flightcrew.io"
	devHostName       = "api.flightcrew.dev"
	devProjectMD5Hash = "fab25c5e9d4d70830074eff996da8ba4"
)

func GetAPIHostName(projectID string, towerName string) string {
	if !strings.Contains(towerName, "dev") {
		return prodHostName
	}

	hash := getMD5Hash(projectID)
	debug.Output("hash is %s", hash)
	if hash == devProjectMD5Hash {
		return devHostName
	}

	return prodHostName
}

func getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
