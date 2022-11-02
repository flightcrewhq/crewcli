package constants

import "fmt"

const (
	ProdBaseURL = "flightcrew.io"
	DevBaseURL  = "flightcrew.dev"

	CLIName = "crewcli"

	appPrefix = "https://app"
	apiPrefix = "api"
)

func GetAppHostName(base string) string {
	return fmt.Sprintf("%s.%s", appPrefix, base)
}

func GetAPIHostName(base string) string {
	return fmt.Sprintf("%s.%s", apiPrefix, base)
}
