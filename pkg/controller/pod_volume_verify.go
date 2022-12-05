package controller

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	_verifyDir     = ".quickon"
	_verifyContent = "This file is auto created in volume backup process."
)

type volumeVerifier struct {
	baseDir    string
	verifyDir  string
	verifyFile string
	logger     logrus.FieldLogger
}

func newVolumeVerifier(volumeDir, backupName string, logger logrus.FieldLogger) *volumeVerifier {
	return &volumeVerifier{
		baseDir:    volumeDir,
		verifyDir:  volumeDir + "/" + _verifyDir,
		verifyFile: _verifyDir + "/" + backupName,

		logger: logger,
	}
}

func (v *volumeVerifier) write() error {
	_, err := os.Stat(v.verifyDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir(v.verifyDir, os.ModePerm); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if err = os.WriteFile(v.baseDir+"/"+v.verifyFile, []byte(_verifyContent), os.FileMode(0644)); err != nil {
		return errors.Wrap(err, "write verify file failed")
	}
	return nil
}

func (v *volumeVerifier) clean() {
	_, err := os.Stat(v.verifyDir)
	if err != nil {
		v.logger.WithError(err).Errorf("detect verify dir %s failed", v.verifyDir)
		return
	}

	err = os.RemoveAll(v.verifyDir)
	if err != nil {
		v.logger.WithError(err).Errorf("remove verify dir %s failed", v.verifyDir)
		return
	}
}
