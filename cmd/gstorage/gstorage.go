package gstorage

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// CopyFile copies files from srcFile to dstFile
//
//	If destinaiton file already exists, it will be overwritten
func CopyFile(srcfile string, dstfile string) error {
	sourcefile, err := os.Open(srcfile)

	if err != nil {
		log.Println("Error reading source file: ", srcfile, err)
		return err
	}
	defer sourcefile.Close()

	destination, err := os.Create(dstfile)

	if err != nil {
		log.Println("Error creating destination file:", destination, err)
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, sourcefile)

	if err != nil {
		log.Println("Error while copying files: ", dstfile, srcfile, err)
		return err
	}

	log.Printf("Successfully copied %s to %s\n", srcfile, dstfile)

	return nil
}

// MoveFile moves files srcfile to dstfile
func MoveFile(srcfile string, dstfile string) error {
	_, err := os.Stat(srcfile)

	if err != nil {
		log.Println("Error reading source file: ", srcfile, err)
		return err
	}

	err = os.Rename(srcfile, dstfile)

	if err != nil {
		log.Println("Error while writing destiation file: ", dstfile, err)
		return err
	}

	log.Printf("Successfully moved %s to %s", srcfile, dstfile)

	return nil
}

// RemoveFile removes/deletes a file or directory
// If srcFile does not exist it returns an error
// If the srcFile is a directory it returns an error
func RemoveFile(srcfile string) error {

	stat, err := os.Stat(srcfile)

	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
	}

	if stat.IsDir() {
		log.Println("Cant remove directory ", srcfile)
		return errors.New("cant remove a directory")
	}

	err = os.Remove(srcfile)

	if err != nil {
		log.Println("Error removing file:", srcfile, err)
		return err
	}
	return nil
}

// ReadFile reads srcfile and it returns its byte size
// It there are errors reading srcfile it returns an error
func ReadFile(srcfile string) ([]byte, error) {
	content, err := os.ReadFile(srcfile)
	if err != nil {
		log.Println("Error readhing file", srcfile)
		return []byte{}, err
	}
	return content, nil
}

func WriteFile(dstFile string, content []byte) error {

	dirpath := filepath.Dir(dstFile)

	err := os.MkdirAll(dirpath, 0755)

	if err != nil {
		log.Println("Unable to create path: ", dirpath)
		return errors.New("unable to create file path")
	}

	file, err := os.Create(dstFile)
	if err != nil {
		log.Panicln("error while creating destination", err)
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	writer.Write(content)
	return nil
}

func ListDir(dirPath string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(dirPath)

	if err != nil {
		return []os.DirEntry{}, err
	}

	return entries, nil
}

func CreateDir(dirPath string, recursive bool) error {
	path := filepath.Dir(dirPath)
	if !recursive {
		_, err := os.ReadDir(path)
		if err != nil {
			log.Println("error creating directory. parent does not exist", path)
			return err
		} else {
			err := os.Mkdir(dirPath, 0755)
			if err != nil {
				log.Println("error creating directory", dirPath, err)
				return err
			}
		}
	} else {
		err := os.MkdirAll(dirPath, 0775)
		if err != nil {
			log.Println("error creating directory", path)
			return err
		}
	}
	log.Printf("successfully created directory %s with recursive %v", dirPath, recursive)
	return nil
}

func RemoveDir(targetDir string) error {

	files, err := os.ReadDir(targetDir)

	if err != nil {
		log.Println("Unable to remove directory.", targetDir, err)
		return err
	}
	if len(files) > 0 {
		log.Println("Unable to delete directory, directory not empty")
		return errors.New("unable to delete directory, directory not empty")
	}
	err = os.Remove(targetDir)
	if err != nil {
		log.Println("Unable to remove directory", targetDir)
		return err
	}
	return nil
}

func RemoveDirAll(targetDir string) error {
	err := os.RemoveAll(targetDir)
	if err != nil {
		log.Println("Unable to remove directory", targetDir, err)
		return err
	}
	return nil
}

func CopyDir(srcDir string, dstDir string) error {
	source, err := os.Stat(srcDir)

	if err != nil {
		log.Println("error occurred while validating", srcDir, err)
		return err
	}
	if !source.IsDir() {
		log.Println("source is not a directory", srcDir)
		return errors.New("source is not a directory")
	}
	destination, err := os.Stat(dstDir)
	if err != nil {

		if err := os.MkdirAll(dstDir, source.Mode()); err != nil {
			log.Println("failed to create destination directory", dstDir)
			return errors.New("failed to create destination directory")
		}
		destination, _ = os.Stat(dstDir)
	}
	if !destination.IsDir() {
		log.Println("destication is not a directory", dstDir)
		return errors.New("destination is not a directory")
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		log.Println("error reading source directory", srcDir, err)
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func FileExists(filename string) (bool, error) {

	_, err := os.Stat(filename)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		log.Println("file does not exist", err)
		return false, nil
	}
	log.Println("file not found", filename, err)
	return false, err
}

func GetFileSize(filename string) (int64, error) {

	stat, err := os.Stat(filename)

	if err == nil {
		return stat.Size(), nil
	}

	if os.IsNotExist(err) {
		log.Println("file does not exist", err)
		return int64(0), err
	}

	return int64(0), err
}

func CalculateFileMD5(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	//Copy the file contents to the hash
	//This streams the file content without loading the entire file
	//which is efficient for large files
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	//Get the byte slice of te hash sum
	// The sum(nil) method computes the final hash and returns is
	hashInBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashInBytes)

	return hashString, nil
}

func CopyFileWithProgress(src, dst string, chunkSize int) (<-chan int64, error) {

	srcFile, err := os.Open(src)

	if err != nil {
		return nil, err
	}

	dstFile, err := os.Create(dst)

	if err != nil {
		srcFile.Close()
		return nil, err
	}

	progressChan := make(chan int64, 10)
	errChan := make(chan error, 1)

	go func() {
		defer srcFile.Close()
		defer dstFile.Close()
		defer close(progressChan)
		defer close(errChan)

		buffer := make([]byte, chunkSize)

		for {
			n, err := srcFile.Read(buffer)
			if n > 0 {
				_, writeErr := dstFile.Write(buffer[:n])
				if writeErr != nil {
					errChan <- errors.New("write failed")
					return
				}
				progressChan <- int64(n)
			}

			if err == io.EOF {
				break
			}
			if err != nil {
				return // Error: just close channel and exit
			}
		}
	}()

	return progressChan, <-errChan
}

type copyJob struct {
	srcPath string
	dstPath string
}

func copyWorker(id int, jobs <-chan copyJob, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		// Copy individual file
		err := CopyFile(job.srcPath, job.dstPath)
		if err != nil {
			errors <- fmt.Errorf("worker %d failed copying %s: %w", id, job.srcPath, err)
			return // Exit on first error
		}
	}
}

func WorkerPoolCopyDir(srcDir, dstDir string, workers int) error {

	srcStat, err := os.Stat(srcDir)

	if err != nil {
		log.Println("error while getting source info", srcDir, err)
		return err
	}

	if !srcStat.IsDir() {
		log.Println("source is not a directory", srcDir, err)
		return errors.New("source is not a directory")
	}

	dstStat, err := os.Stat(dstDir)

	if err != nil {
		if os.IsNotExist(err) {
			// If dst does not exit, create it
			log.Println("destination does not exist. creating")
			err := os.MkdirAll(dstDir, srcStat.Mode())
			if err != nil {
				log.Println("error while creating destination directory", dstDir, err)
				return err
			}
			log.Println("created destination directory")
		}
		// Pass any other error
		log.Println("error while getting destination info", err)
		return err
	}

	if !dstStat.IsDir() {
		return errors.New("destination is not a directory")
	}

	// Create directory structure frist

	// HINT: Create all directories BEFORE starting file workers
	// This avoids race conditions where workers try to copy files
	// to directories that don't exist yet

	walkErr := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Calculate relative path and create in destination
			relPath, _ := filepath.Rel(srcDir, path)
			dstPath := filepath.Join(dstDir, relPath)
			return os.MkdirAll(dstPath, 0755)
		}
		return nil
	})

	if walkErr != nil {
		log.Println("error while creating directory structure:", walkErr)
		if removeErr := os.RemoveAll(dstDir); removeErr != nil {
			log.Println("failed to clean up destination:", removeErr)
		}
		return walkErr // <-- Return original error, not cleanup error
	}

	var wg sync.WaitGroup
	jobQueue := make(chan copyJob, 100) // Buffer jobs
	errorChan := make(chan error, 1)    // Collect errors

	// start worker pool
	for i := 1; i <= workers; i++ {
		wg.Add(1)
		go copyWorker(i, jobQueue, errorChan, &wg)
	}

	// Send only FILE jobs to workers (directories already created)
	walkErr = filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() { // Only send files
			relPath, _ := filepath.Rel(srcDir, path)
			dstPath := filepath.Join(dstDir, relPath)

			jobQueue <- copyJob{
				srcPath: path,
				dstPath: dstPath,
			}
		}
		return nil
	})

	close(jobQueue)
	wg.Wait()
	close(errorChan)

	// Check walkErr first
	if walkErr != nil {
		log.Println("error while walking directory:", walkErr)
		return walkErr
	}

	// Then check worker errors
	select {
	case err := <-errorChan:
		return err
	default:
		return nil
	}
}
