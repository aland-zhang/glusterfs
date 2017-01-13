package controller

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/appscode/log"
)

func firstController(m map[string]string) (bool, *GlusterControllerSpec) {
	val, ok := m[glusterFSClusterControllerData]
	log.Infoln("first controller status --", !ok)
	if ok {
		specs := &GlusterControllerSpec{}
		err := json.Unmarshal([]byte(val), specs)
		if err == nil {
			return false, specs
		}
	}
	return true, nil
}

func newStringReader(ss []string) io.Reader {
	formattedString := strings.Join(ss, "\n")
	reader := strings.NewReader(formattedString)
	return reader
}
