package cmd

import "os"

func GetGFIDDir() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return os.TempDir() + "/" + hostname + ".gfid", nil
}
