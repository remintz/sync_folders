package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	// Ensure correct number of arguments
	if len(os.Args) != 3 {
		fmt.Println("Usage: sync_folders <source folder> <destination folder>")
		return
	}

	folderA := os.Args[1]
	folderB := os.Args[2]

	// Synchronize folders
	err := syncFolders(folderA, folderB)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Synchronization between %s and %s completed.\n", folderA, folderB)
}

func syncFolders(folderA, folderB string) error {
	// Synchronize from A to B
	if err := syncFromTo(folderA, folderB); err != nil {
		return err
	}
	// Synchronize from B to A
	if err := syncFromTo(folderB, folderA); err != nil {
		return err
	}
	return nil
}

func syncFromTo(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if _, err := os.Stat(dstPath); os.IsNotExist(err) {
				// Create the directory and all necessary parents
				if err := os.MkdirAll(dstPath, 0755); err != nil {
					return err
				}
				fmt.Printf("Directory created: %s\n", dstPath)
			}
			if err := syncFromTo(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := syncFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func syncFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	dstInfo, err := os.Stat(dst)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}
		if err := copyFile(src, dst, srcInfo); err != nil {
			return err
		}
		fmt.Printf("File created: %s\n", dst)
		return nil
	} else if err != nil {
		return err
	}

	if !sameFile(srcInfo, dstInfo) {
		if srcInfo.ModTime().After(dstInfo.ModTime()) {
			if err := copyFile(src, dst, srcInfo); err != nil {
				return err
			}
			fmt.Printf("File updated: %s\n", dst)
		} else {
			if err := copyFile(dst, src, dstInfo); err != nil {
				return err
			}
			fmt.Printf("File updated: %s\n", src)
		}
	}
	return nil
}

func sameFile(srcInfo, dstInfo os.FileInfo) bool {
	return srcInfo.Size() == dstInfo.Size() && srcInfo.ModTime().Equal(dstInfo.ModTime())
}

func copyFile(src, dst string, info os.FileInfo) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	if _, err := io.Copy(destinationFile, sourceFile); err != nil {
		return err
	}

	if err := os.Chtimes(dst, info.ModTime(), info.ModTime()); err != nil {
		return err
	}
	return nil
}
