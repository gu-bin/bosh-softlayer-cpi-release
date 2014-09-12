package disk

import (
	"path/filepath"
	"strconv"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

const fsCreatorLogTag = "FSCreator"

type FSCreator struct {
	dirPath string

	fs        boshsys.FileSystem
	cmdRunner boshsys.CmdRunner
	logger    boshlog.Logger
}

func NewFSCreator(
	dirPath string,
	fs boshsys.FileSystem,
	cmdRunner boshsys.CmdRunner,
	logger boshlog.Logger,
) FSCreator {
	return FSCreator{
		dirPath:   dirPath,
		fs:        fs,
		cmdRunner: cmdRunner,
		logger:    logger,
	}
}

func (c FSCreator) Create(size int) (Disk, error) {
	c.logger.Debug(fsCreatorLogTag, "Creating disk of size '%d'", size)

	id := 1234 //TODO: get disk ID from SL API

	diskPath := filepath.Join(c.dirPath, strconv.Itoa(id))

	err := c.fs.WriteFile(diskPath, []byte{})
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating empty disk")
	}

	sizeStr := strconv.Itoa(20) + "MB"

	_, _, _, err = c.cmdRunner.RunCommand("truncate", "-s", sizeStr, diskPath)
	if err != nil {
		c.cleanUpFile(diskPath)
		return nil, bosherr.WrapError(err, "Resizing disk to '%s'", sizeStr)
	}

	_, _, _, err = c.cmdRunner.RunCommand("/sbin/mkfs", "-t", "ext4", "-F", diskPath)
	if err != nil {
		c.cleanUpFile(diskPath)
		return nil, bosherr.WrapError(err, "Building disk filesystem '%s'", diskPath)
	}

	return NewFSDisk(id, diskPath, c.fs, c.logger), nil
}

func (c FSCreator) cleanUpFile(path string) {
	err := c.fs.RemoveAll(path)
	if err != nil {
		c.logger.Error(fsCreatorLogTag, "Failed deleting file '%s': %s", path, err.Error())
	}
}