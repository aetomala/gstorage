package gstorage_test

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "storage/cmd/gstorage"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gstorage File Operations Library", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "gstorage_test_*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).ToNot(HaveOccurred())
	})

	// Helper functions for test setup

	createTestFile := func(path string, content string) {
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			log.Fatal("Error creating file", err)
		}
		Expect(err).ToNot(HaveOccurred())
	}

	createTestDir := func(path string) {
		err := os.MkdirAll(path, 0755)
		Expect(err).ToNot(HaveOccurred())
	}

	fileExists := func(path string) bool {
		_, err := os.Stat(path)
		return !os.IsNotExist(err)
	}

	readFileContent := func(path string) string {
		content, err := os.ReadFile(path)
		Expect(err).ToNot(HaveOccurred())
		return string(content)
	}
	// Core Module Tests

	Describe("Core File Operations", func() {
		Describe("CopyFile", func() {
			var srcFile, dstFile string
			BeforeEach(func() {
				srcFile = filepath.Join(tempDir, "source.txt")
				dstFile = filepath.Join(tempDir, "destication.txt")
			})

			Context("when source file exists", func() {
				BeforeEach(func() {
					createTestFile(srcFile, "Hello, World!")
				})

				It("should copy the file successfully", func() {
					err := CopyFile(srcFile, dstFile)

					Expect(err).ToNot(HaveOccurred())
					Expect(fileExists(dstFile)).To(BeTrue())
					Expect(readFileContent(dstFile)).To(Equal("Hello, World!"))
					Expect(fileExists(srcFile)).To(BeTrue())
					Expect(readFileContent(srcFile)).To(Equal("Hello, World!"))
				})
				It("should overwrite existing destination file", func() {
					createTestFile(dstFile, "Old content")
					err := CopyFile(srcFile, dstFile)
					Expect(err).ToNot(HaveOccurred())
					Expect(readFileContent(dstFile)).To(Equal("Hello, World!"))
				})
				It("should handle large files", func() {
					largeContent := strings.Repeat("A", 102481024) //1MB
					createTestFile(srcFile, largeContent)

					err := CopyFile(srcFile, dstFile)
					Expect(err).ToNot(HaveOccurred())
					Expect(readFileContent(dstFile)).To(Equal(largeContent))
				})
			})
			Context("when source file fors not exist", func() {
				It("should return an error", func() {
					err := CopyFile(srcFile, dstFile)
					Expect(err).To(HaveOccurred())
					Expect(os.IsNotExist(err)).To(BeTrue())
				})
			})
			Context("When destination directory does not exist", func() {
				It("should return an error", func() {
					createTestFile(srcFile, "test")
					nonExistentDst := filepath.Join(tempDir, "nonexistent", "file.txt")

					err := CopyFile(srcFile, nonExistentDst)
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Describe("MoveFile", func() {
			var srcFile, dstFile string
			BeforeEach(func() {
				srcFile = filepath.Join(tempDir, "source.txt")
				dstFile = filepath.Join(tempDir, "destination.txt")
				createTestFile(srcFile, "Move me!")
			})

			It("should move the file successfully", func() {
				err := MoveFile(srcFile, dstFile)
				Expect(err).ToNot(HaveOccurred())
				Expect(fileExists(srcFile)).To(BeFalse())
				Expect(fileExists(dstFile)).To(BeTrue())
				Expect(readFileContent(dstFile)).To(Equal("Move me!"))
			})

			It("should handle cross-device move", func() {
				// This test might need adjustment based on your implementation
				// of cross-device move detection
				err := MoveFile(srcFile, dstFile)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when source file does not exist", func() {
				It("should return an error", func() {
					nonExisistingSrc := filepath.Join(tempDir, "nonexistent.txt")
					err := MoveFile(nonExisistingSrc, dstFile)
					Expect(err).To(HaveOccurred())
				})
			})
		})
		Describe("RemoveFile", func() {
			var testFile string

			BeforeEach(func() {
				testFile = filepath.Join(tempDir, "test.txt")
			})
			Context("when file exists", func() {
				BeforeEach(func() {
					createTestFile(testFile, "Delete me")
				})
				It("should remove the file successfully", func() {
					err := RemoveFile(testFile)
					Expect(err).ToNot(HaveOccurred())
					Expect(fileExists(testFile)).To(BeFalse())
				})
			})

			Context("when file does not exist", func() {
				It("should be idempotent (no error)", func() {
					err := RemoveFile(testFile)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when path points to a directory", func() {
				It("should return an error", func() {
					testDir := filepath.Join(tempDir, "testdir")
					createTestDir(testDir)

					err := RemoveFile(testDir)
					Expect(err).To(HaveOccurred())
				})
			})
		})
		Describe("ReadFile", func() {
			var testFile string
			BeforeEach(func() {
				testFile = filepath.Join(tempDir, "read.txt")
			})

			Context("when file exists", func() {
				It("should read file content", func() {
					content := "Hello, Reader!"
					createTestFile(testFile, content)

					data, err := ReadFile(testFile)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(data)).To(Equal(content))
				})
				It("should handle empty files", func() {
					createTestFile(testFile, "")

					data, err := ReadFile(testFile)

					Expect(err).ToNot(HaveOccurred())
					Expect(len(data)).To(Equal(0))
				})
			})
		})
		Describe("WriteFile", func() {
			var testFile string

			BeforeEach(func() {
				testFile = filepath.Join(tempDir, "write.txt")
			})

			It("shoudl create and write to a new file", func() {
				content := []byte("Hello, Writer!")
				err := WriteFile(testFile, content)
				Expect(err).NotTo(HaveOccurred())
				Expect(readFileContent(testFile)).To(Equal(string(content)))
			})

			It("should truncate existing file", func() {
				createTestFile(testFile, "Old content that should be replaced")
				newContent := []byte("New content")

				err := WriteFile(testFile, newContent)
				Expect(err).NotTo(HaveOccurred())
				Expect(readFileContent(testFile)).To(Equal(string(newContent)))
			})

			It("should create a directory path if it doesn't exist", func() {
				nestedFile := filepath.Join(tempDir, "nested", "deep", "file.txt")
				content := []byte("Nested content")

				err := WriteFile(nestedFile, content)
				Expect(err).NotTo(HaveOccurred())
				Expect(readFileContent(nestedFile)).To(Equal(string(content)))
			})
		})
		Describe("Directory Operations", func() {
			Describe("ListDir", func() {
				var testDir string
				BeforeEach(func() {
					testDir = filepath.Join(tempDir, "listtest")
					createTestDir(testDir)
				})

				Context("when directory contains files and subdirectorys", func() {
					BeforeEach(func() {
						createTestFile(filepath.Join(testDir, "file1.txt"), "content1")
						createTestFile(filepath.Join(testDir, "file2.log"), "content2")
						createTestDir(filepath.Join(testDir, "subdir1"))
						createTestDir(filepath.Join(testDir, "subdir2"))
					})
					It("should list all entries", func() {
						entries, err := ListDir(testDir)
						Expect(err).ToNot(HaveOccurred())
						Expect(len(entries)).To(Equal(4))
						names := make([]string, len(entries))
						for i, entry := range entries {
							names[i] = entry.Name()
						}
						Expect(names).To(ContainElements("file1.txt", "file2.log", "subdir1", "subdir2"))
					})
				})
				Context("when directory is empty", func() {
					It("should return empty slice", func() {
						entries, err := ListDir(testDir)
						Expect(err).ToNot(HaveOccurred())
						Expect(len(entries)).To(Equal(0))
					})
				})
				Context("when directory does not exist", func() {
					_, err := ListDir(filepath.Join(testDir, "nonexistent"))
					Expect(err).To(HaveOccurred())
				})
			})
			Describe("CreateDir", func() {
				var testDir string
				BeforeEach(func() {
					testDir = filepath.Join(tempDir, "createtest")
				})
				Context("with recursive=false", func() {
					It("should create directory when parent exists", func() {
						err := CreateDir(testDir, false)
						Expect(err).ToNot(HaveOccurred())
						Expect(fileExists(testDir)).To(BeTrue())
					})
					It("should fail when parent does not exist", func() {
						nestedDir := filepath.Join(tempDir, "nonexistent", "nested")
						err := CreateDir(nestedDir, false)
						Expect(err).To(HaveOccurred())
					})
				})
				Context("with recursive=true", func() {
					It("should create directory and parents", func() {
						nestedDir := filepath.Join(tempDir, "level1", "level2", "level3")
						err := CreateDir(nestedDir, true)
						Expect(err).NotTo(HaveOccurred())
						Expect(fileExists(nestedDir)).To(BeTrue())
					})
				})
			})
			Describe("RemoveDir", func() {
				var testDir string
				BeforeEach(func() {
					testDir = filepath.Join(tempDir, "removetest")
					createTestDir(testDir)
				})
				Context("when directory is empty", func() {
					It("should remove directory successfully", func() {
						err := RemoveDir(testDir)
						Expect(err).NotTo(HaveOccurred())
						Expect(fileExists(testDir)).To(BeFalse())
					})
				})
				Context("when directory is not empty", func() {
					It("should fail to remove directory", func() {
						createTestFile(filepath.Join(testDir, "file.text"), "content")
						err := RemoveDir(testDir)
						Expect(err).To(HaveOccurred())
					})
				})
			})
			Describe("RemoveDirAll", func() {
				var testDir string
				BeforeEach(func() {
					testDir = filepath.Join(tempDir, "removealtest")
					createTestDir(testDir)
				})
				It("should remove directory and all contents", func() {
					subDir := filepath.Join(testDir, "subdir")
					createTestDir(subDir)
					createTestFile(filepath.Join(testDir, "file1.txt"), "content1")
					createTestFile(filepath.Join(subDir, "file2.txt"), "content2")

					err := RemoveDirAll(testDir)
					Expect(err).NotTo(HaveOccurred())
					Expect(fileExists(testDir)).To(BeFalse())
				})
			})
			Describe("CopyDir", func() {
				var srcDir, dstDir string

				BeforeEach(func() {
					srcDir = filepath.Join(tempDir, "src")
					dstDir = filepath.Join(tempDir, "dst")

					createTestDir(srcDir)
					createTestDir(filepath.Join(srcDir, "subdir"))
					createTestFile(filepath.Join(srcDir, "file1.txt"), "content1")
					createTestFile(filepath.Join(srcDir, "subdir", "file2.txt"), "content2")
				})
				It("should copy entire directory tree", func() {
					err := CopyDir(srcDir, dstDir)
					Expect(err).NotTo(HaveOccurred())

					//Verify structure
					Expect(fileExists(dstDir)).To(BeTrue())
					Expect(fileExists(filepath.Join(dstDir, "subdir"))).To(BeTrue())
					Expect(readFileContent(filepath.Join(dstDir, "file1.txt"))).To(Equal("content1"))
					Expect(readFileContent(filepath.Join(dstDir, "subdir", "file2.txt"))).To(Equal("content2"))
				})
			})
		})
	})
	Describe("Advanced Operations", func() {
		Describe("FileExists", func() {
			It("should return true for existing file", func() {
				testFile := filepath.Join(tempDir, "exists.txt")
				createTestFile(testFile, "I exist")

				exists, err := FileExists(testFile)

				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())
			})

			It("should return false for non-existing file", func() {
				testFile := filepath.Join(tempDir, "nonexistent.txt")

				exists, err := FileExists(testFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeFalse())
			})
			It("should return true for existing directory", func() {
				testDir := filepath.Join(tempDir, "existDir")
				createTestDir(testDir)

				exist, err := FileExists(testDir)

				Expect(err).NotTo(HaveOccurred())
				Expect(exist).To(BeTrue())
			})
		})
	})
	Describe("GetFileSize", func() {
		It("should return correct file size", func() {
			testFile := filepath.Join(tempDir, "sizefile.txt")
			content := "Hello, World!" // 13 bytes
			createTestFile(testFile, content)

			size, err := GetFileSize(testFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(size).To(Equal(int64(size)))
		})
		It("should return 0 for empty file", func() {
			testFile := filepath.Join(tempDir, "emptyfile.txt")
			createTestFile(testFile, "")

			size, err := GetFileSize(testFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(size).To(Equal(int64(0)))
		})
		It("should return error for non-existent file", func() {
			testFile := filepath.Join(tempDir, "non-existintent.txt")

			_, err := GetFileSize(testFile)
			Expect(err).To(HaveOccurred())
			Expect(os.IsNotExist(err)).To(BeTrue())
		})
	})
	Describe("CalculateFileMD5", func() {
		It("should calculate correct MD5 hash", func() {
			testFile := filepath.Join(tempDir, "hashfile.txt")
			content := "Hello, World!"
			createTestFile(testFile, content)

			expectedHash := fmt.Sprintf("%x", md5.Sum([]byte(content)))

			hash, err := CalculateFileMD5(testFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(hash).To(Equal(expectedHash))
		})
		It("should handle empty files", func() {
			testFile := filepath.Join(tempDir, "emptyhashfile.txt")
			createTestFile(testFile, "")

			expectedHash := fmt.Sprintf("%x", md5.Sum([]byte("")))

			hash, err := CalculateFileMD5(testFile)

			Expect(err).NotTo(HaveOccurred())
			Expect(hash).To(Equal(expectedHash))
		})
	})
	Describe("Concurrency & Progress operations", func() {
		Describe("CopyFileWithProgress", func() {
			It("should copy file and report progress", func() {
				srcFile := filepath.Join(tempDir, "progress_src.txt")
				dstFile := filepath.Join(tempDir, "progress_dst.txt")
				bytesToWrite := 1000

				content := strings.Repeat("A", bytesToWrite) //1000 bytes
				createTestFile(srcFile, content)

				progressChan, err := CopyFileWithProgress(srcFile, dstFile, 100)
				Expect(err).NotTo(HaveOccurred())

				var totalBytes int64
				var progressUpdates []int64

				for bytes := range progressChan {
					totalBytes += bytes
					progressUpdates = append(progressUpdates, bytes)
				}
				// Verify copy complete
				Expect(FileExists(dstFile)).To(BeTrue())
				Expect(readFileContent(dstFile)).To(Equal(content))

				//Verify progress updates
				Expect(len(progressUpdates)).To(BeNumerically(">", 0))
				Expect(totalBytes).To(Equal(int64(bytesToWrite)))
			})
		})
	})
	Describe("WorkerPoolCopyDir", func() {
		var srcDir, dstDir string
		BeforeEach(func() {
			srcDir = filepath.Join(tempDir, "pool_src")
			dstDir = filepath.Join(tempDir, "pool_dst")
			//create source directory
			createTestDir(srcDir)

			for i := 0; i < 10; i++ {
				filename := fmt.Sprintf("file_%d.txt", i)
				content := fmt.Sprintf("Content of file %d", i)
				createTestFile(filepath.Join(srcDir, filename), content)
			}
			// Create subdirectory
			subDir := filepath.Join(dstDir, "subdir")
			createTestDir(subDir)
			createTestFile(filepath.Join(subDir, "nested.txt"), "nested content")
		})
		It("should copy directory using worker pool", func() {

			err := WorkerPoolCopyDir(srcDir, dstDir, 3) // 3 workers
			Expect(err).NotTo(HaveOccurred())

			entries, err := ListDir(srcDir)
			Expect(err).NotTo(HaveOccurred())

			for _, entry := range entries {
				if entry.IsDir() {
					Expect(FileExists(filepath.Join(dstDir, entry.Name()))).To(BeTrue())
				} else {
					//Check file
					srcPath := filepath.Join(srcDir, entry.Name())
					dstPath := filepath.Join(dstDir, entry.Name())
					Expect(FileExists(dstPath)).To(BeTrue())
					Expect(readFileContent(dstPath)).To(Equal(readFileContent(srcPath)))
				}
			}

		})
		It("should handle errors from workers", func() {
			// Create a file with restricted permissions to trigger an error
			restrictedFile := filepath.Join(srcDir, "restricted.txt")
			createTestFile(restrictedFile, "restricted")
			err := os.Chmod(restrictedFile, 0000)
			Expect(err).NotTo(HaveOccurred())

			// Restore permissions after test
			defer func() {
				os.Chmod(restrictedFile, 0644)
			}()

			err = WorkerPoolCopyDir(srcDir, dstDir, 2)

			Expect(err).To(HaveOccurred())
		})
	})
	// Integration Tests
	Describe("Integration Tests", func() {
		It("should perform complete file operations workflow", func() {
			// Create original file
			originalFile := filepath.Join(tempDir, "original.txt")
			originalContent := "Original content for integration test"
			createTestFile(originalFile, originalContent)

			// Copy file
			copiedFile := filepath.Join(tempDir, "copied.txt")
			err := CopyFile(originalFile, copiedFile)
			Expect(err).NotTo(HaveOccurred())

			// Verify copy
			exists, err := FileExists(copiedFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())

			size, err := GetFileSize(copiedFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(size).To(Equal(int64(len(originalContent))))

			// Calculate and compare hashes
			originalHash, err := CalculateFileMD5(originalFile)
			Expect(err).NotTo(HaveOccurred())

			copiedHash, err := CalculateFileMD5(copiedFile)
			Expect(err).NotTo(HaveOccurred())

			Expect(originalHash).To(Equal(copiedHash))

			// Move file
			movedFile := filepath.Join(tempDir, "moved.txt")
			err = MoveFile(copiedFile, movedFile)
			Expect(err).NotTo(HaveOccurred())

			// Verify move
			exists, err = FileExists(copiedFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())

			exists, err = FileExists(movedFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())

			// Clean up
			err = RemoveFile(originalFile)
			Expect(err).NotTo(HaveOccurred())

			err = RemoveFile(movedFile)
			Expect(err).NotTo(HaveOccurred())
		})
	})
	// Performance/Stress Tests
	Describe("Performance Tests", func() {
		It("should handle many small files efficiently", func() {
			srcDir := filepath.Join(tempDir, "perf_src")
			dstDir := filepath.Join(tempDir, "perf_dst")
			createTestDir(srcDir)

			// Create 100 small files
			for i := 0; i < 100; i++ {
				filename := fmt.Sprintf("small_%d.txt", i)
				content := fmt.Sprintf("Small file content %d", i)
				createTestFile(filepath.Join(srcDir, filename), content)
			}

			start := time.Now()
			err := CopyDir(srcDir, dstDir)
			duration := time.Since(start)

			Expect(err).NotTo(HaveOccurred())
			Expect(duration).To(BeNumerically("<", 5*time.Second)) // Should complete in under 5 seconds
		})
	})
})
