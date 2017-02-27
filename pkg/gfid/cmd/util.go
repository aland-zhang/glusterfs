package cmd

import "os"

func GetStorageDir() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return os.TempDir() + "/" + hostname + ".gfid", nil
}
