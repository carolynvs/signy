package trust

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/cnabio/signy/pkg/docker"
	"github.com/cnabio/signy/pkg/intoto"
	"github.com/cnabio/signy/pkg/tuf"
)

// ValidateThinBundle runs the TUF and in-toto validations for a CNAB bundle in thin format (canonical JSON form)
func ValidateThinBundle(ref, trustServer, tlscacert, trustDir, verificationImage, logLevel, timeout string, targets []string, keep bool) error {
	err := tuf.VerifyTrust(ref, "", trustServer, tlscacert, trustDir, timeout)
	if err != nil {
		return err
	}

	return runVerifications(ref, trustServer, tlscacert, trustDir, verificationImage, logLevel, timeout, targets, keep)
}

// ValidateThickBundle runs the TUF and in-toto validations for a CNAB bundle in thick format
func ValidateThickBundle(ref, file, trustServer, tlscacert, trustDir, verificationImage, logLevel, timeout string, targets []string, keep bool) error {
	err := tuf.VerifyTrust(ref, file, trustServer, tlscacert, trustDir, timeout)
	if err != nil {
		return err
	}

	return runVerifications(ref, trustServer, tlscacert, trustDir, verificationImage, logLevel, timeout, targets, keep)
}

func runVerifications(ref, trustServer, tlscacert, trustDir, verificationImage, logLevel, timeout string, targets []string, keep bool) error {
	target, _, err := tuf.GetTargetAndSHA(ref, trustServer, tlscacert, trustDir, timeout)
	if err != nil {
		return err
	}

	m := &intoto.Metadata{}
	err = json.Unmarshal(*target.Custom, m)
	if err != nil {
		return err
	}

	verificationDir, err := ioutil.TempDir(os.TempDir(), "intoto-verification")
	if err != nil {
		return err
	}
	if !keep {
		defer func() {
			os.RemoveAll(verificationDir)
			os.Remove(verificationDir)
		}()
	}

	log.Infof("Writing In-Toto metadata files into %v", verificationDir)
	err = intoto.WriteMetadataFiles(m, verificationDir)
	if err != nil {
		return err
	}

	return docker.Run(verificationImage, filepath.Join(verificationDir, intoto.LayoutDefaultName), filepath.Join(verificationDir, intoto.KeyDefaultName), verificationDir, logLevel, targets)
}
