package gcp

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"flightcrew.io/cli/internal/constants"
	"flightcrew.io/cli/internal/debug"
)

const (
	devProjectMD5Hash = "fab25c5e9d4d70830074eff996da8ba4"
)

func GetHostBaseURL(projectID string, towerName string) string {
	if !strings.Contains(towerName, "dev") {
		return constants.ProdBaseURL
	}

	hash := getMD5Hash(projectID)
	debug.Output("hash is %s", hash)
	if hash == devProjectMD5Hash {
		return constants.DevBaseURL
	}

	return constants.ProdBaseURL
}

func getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
