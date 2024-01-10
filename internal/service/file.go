package service

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/pkg/randstring"
)

type FileService struct {
}

func NewFileService() *FileService {
	return &FileService{}
}

func (s *FileService) GetFiles(path string) ([]fs.FileInfo, error) {
	fileList, err := ioutil.ReadDir("../../storage" + path)
	if err != nil {
		return fileList, errors.New("path not found")
	}

	return fileList, nil
}

func (s *FileService) DeleteFiles(path string) error {
	if path == "" {
		path = "/"
	}

	fullPath := "../../storage" + path

	err := os.RemoveAll(fullPath)
	if err != nil {
		return errors.New("failed to delete file or directory")
	}

	return nil
}

func (s *FileService) CreateFolder(name string, path string) error {
	fullPath := fmt.Sprintf("../../storage%s/%s", path, name)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		err := os.MkdirAll(fullPath, os.ModePerm)
		if err != nil {
			log.Println(err)
			return err
		}
	} else {
		fullPath = fullPath + "-" + randstring.RandAlphanumString(5)
		err := os.MkdirAll(fullPath, os.ModePerm)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

// Rename or move file.
func (s *FileService) RenameFile(fullPath string, newFullPath string) error {
	err := os.Rename("../../storage"+fullPath, "../../storage"+newFullPath)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
