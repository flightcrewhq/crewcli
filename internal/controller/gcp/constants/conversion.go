package constants

var (
	FlagToKey = map[string]string{
		FlagProject:        KeyProject,
		FlagVirtualMachine: KeyVirtualMachine,
		FlagPlatform:       KeyPlatform,
		FlagToken:          KeyAPIToken,
		FlagTowerVersion:   KeyTowerVersion,
		FlagZone:           KeyZone,
		FlagWrite:          KeyPermissions,
	}
)
