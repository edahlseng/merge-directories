package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func mergeDirectories(moveInsteadOfCopy bool, sourceDirectory string, destinationDirectory string) bool {
	if sourceDirectory == destinationDirectory {
		return false
	}

	encounteredConflicts := false

	err := filepath.Walk(sourceDirectory, func(dirtyRelativePath string, info os.FileInfo, err error) error {
		if path.Clean(dirtyRelativePath) == path.Clean(sourceDirectory) {
			return nil
		}
		relativePath := strings.TrimPrefix(path.Clean(dirtyRelativePath), path.Clean(sourceDirectory) + "/")
		if ! info.IsDir() {
			sourcePath := path.Join(sourceDirectory, relativePath)
			destinationPath := path.Join(destinationDirectory, relativePath)

			if _, err := os.Stat(destinationPath); !os.IsNotExist(err) {
				sourceFile, err := os.Open(sourcePath)
				if err != nil {
					fmt.Println("Encountered error while comparing", sourcePath, "and", destinationPath, ":", err)
					return err
				}
				defer sourceFile.Close()

				sourceHash := md5.New()
				_, err = io.Copy(sourceHash, sourceFile);
				if err != nil {
					fmt.Println("Encountered error while comparing", sourcePath, "and", destinationPath, ":", err)
					return err
				}

				destinationFile, err := os.Open(destinationPath)
				if err != nil {
					fmt.Println("Encountered error while comparing", sourcePath, "and", destinationPath, ":", err)
					return err
				}
				defer destinationFile.Close()

				destinationHash := md5.New()
				_, err = io.Copy(destinationHash, destinationFile);
				if err != nil {
					fmt.Println("Encountered error while comparing", sourcePath, "and", destinationPath, ":", err)
					return err
				}

				// TODO: could just compare the two read streams rather than hashing and then comparing the hashes. More work for the future...
				if bytes.Compare(sourceHash.Sum(nil)[:16], destinationHash.Sum(nil)[:16]) != 0 {
					encounteredConflicts = true
					return nil
				}
			}

			err := os.MkdirAll(path.Dir(destinationPath), os.ModePerm)
			if err != nil {
				fmt.Println("Encountered error while trying to create destination paths", sourcePath, "and", destinationPath, ":", err)
				return err
			}

			if moveInsteadOfCopy {
				err := os.Rename(sourcePath, destinationPath)
				if err != nil {
					fmt.Println("Encountered error moving", sourcePath, "and", destinationPath, ":", err)
					return err
				}
			} else {
				from, err := os.Open(sourcePath)
				if err != nil {
					fmt.Println("Encountered error copying", sourcePath, "and", destinationPath, ":", err)
					return err
				}
				defer from.Close()

				to, err := os.OpenFile(destinationPath, os.O_RDWR|os.O_CREATE, info.Mode().Perm())
				if err != nil {
					fmt.Println("Encountered error copying", sourcePath, "and", destinationPath, ":", err)
					return err
				}
				defer to.Close()

				_, err = io.Copy(to, from)
				if err != nil {
					fmt.Println("Encountered error copying", sourcePath, "and", destinationPath, ":", err)
					return err
				}
			}
		}
		return nil
	})

	if err != nil || encounteredConflicts {
		return true
	}

	return false
}

func main() {
	moveInsteadOfCopy := false
	sourceDirectory := ""
	destinationDirectory := ""

	argumentsWithoutProgram := os.Args[1:]

	for _, element := range argumentsWithoutProgram {
		if element == "--mv" {
			moveInsteadOfCopy = true
			continue
		}

		if sourceDirectory == "" {
			sourceDirectory = element
			continue
		}

		if destinationDirectory == "" {
			destinationDirectory = element
			continue
		}

		fmt.Println("Unknown option", element)
		os.Exit(1)
	}

	if _, err := os.Stat(sourceDirectory); os.IsNotExist(err) {
		fmt.Println("Source directory does not exist")
		os.Exit(1)
	}

	if _, err := os.Stat(destinationDirectory); os.IsNotExist(err) {
		fmt.Println("Destination directory does not exist")
		os.Exit(1)
	}

	encounteredConflicts := mergeDirectories(moveInsteadOfCopy, sourceDirectory, destinationDirectory)
	if encounteredConflicts {
		fmt.Println("Encountered conflicting files or error")
		os.Exit(1)
	}
}
