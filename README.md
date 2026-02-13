# gstorage

![Tests](https://github.com/aetomala/gstorage/actions/workflows/test.yml/badge.svg?branch=main)
![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

**Production-grade file and directory operations library - exploring Go filesystem patterns and concurrent I/O abstractions**

## Purpose

Part of my ongoing platform engineering skill maintenance. This repo explores implementing a comprehensive file storage operations library - the kind of utilities that appear in every infrastructure tool, container runtime, and deployment system.

Built incrementally using TDD (Ginkgo/Gomega) to practice Go's filesystem APIs, error handling patterns, concurrent operations, and abstraction design for common I/O tasks.

## What This Implements

A complete file and directory operations library demonstrating:

**Core patterns**:
- Single file operations (copy, move, delete, read, write)
- Directory traversal and recursive operations
- Metadata queries (file exists, size, hash)
- Worker pool pattern for parallel operations
- Progress tracking for long-running operations
- Efficient streaming for large files

**Architecture**:
- Clean, focused API for common filesystem tasks
- Goroutine-based worker pool for concurrent copies
- Channel-based progress reporting
- Proper error handling with context
- Buffered channels for job queueing

**Use case**: Foundation for backup systems, file synchronizers, deployment tools, and container runtimes that need reliable file operations at scale.

## Architecture

**Key Functions**:
```go
// Single file operations
func CopyFile(srcfile string, dstfile string) error
func MoveFile(srcfile string, dstfile string) error
func RemoveFile(srcfile string) error
func ReadFile(srcfile string) ([]byte, error)
func WriteFile(dstFile string, content []byte) error

// Directory operations
func ListDir(dirPath string) ([]os.DirEntry, error)
func CreateDir(dirPath string, recursive bool) error
func RemoveDir(targetDir string) error
func RemoveDirAll(targetDir string) error
func CopyDir(srcDir string, dstDir string) error

// Metadata operations
func FileExists(filename string) (bool, error)
func GetFileSize(filename string) (int64, error)
func CalculateFileMD5(filename string) (string, error)

// Advanced operations
func CopyFileWithProgress(src, dst string, chunkSize int) (<-chan int64, error)
func WorkerPoolCopyDir(srcDir, dstDir string, workers int) error
```

**Worker Pool Pattern** (for large directory copies):
```go
// Splits work into two phases:
// 1. Directory structure creation (serial, ensures no races)
// 2. File copying (parallel, via worker pool)
//
// This avoids race conditions where workers try to copy files
// to directories that don't exist yet

type copyJob struct {
    srcPath string
    dstPath string
}

func copyWorker(id int, jobs <-chan copyJob, errors chan<- error, wg *sync.WaitGroup) {
    defer wg.Done()
    for job := range jobs {
        err := CopyFile(job.srcPath, job.dstPath)
        if err != nil {
            errors <- fmt.Errorf("worker %d failed: %w", id, job.srcPath, err)
            return
        }
    }
}
```

**Progress Tracking** (for monitoring operations):
```go
progressChan, err := CopyFileWithProgress(src, dst, 64*1024) // 64KB chunks
for bytes := range progressChan {
    totalCopied += bytes
    fmt.Printf("Copied %d bytes\n", totalCopied)
}
```

## Why This Pattern Matters

File operations are deceptively complex in production systems:

- **Correctness**: File copies must preserve permissions, handle permissions properly, work across filesystems
- **Efficiency**: Large file copies should stream, not load entirely into memory
- **Concurrency**: Recursive directory operations benefit from worker pools
- **Observability**: Progress reporting matters for large operations
- **Error recovery**: Operations may need cleanup on partial failure
- **Idempotency**: Some operations (RemoveFile) should be idempotent

This appears in every production tool: rsync, Docker, Kubernetes, Terraform, cloud CLIs.

## Concurrency Patterns Explored

**Worker Pool for Parallel Operations**:
```go
// Multiple workers process jobs from shared channel
var wg sync.WaitGroup
jobQueue := make(chan copyJob, 100)
errorChan := make(chan error, 1)

for i := 1; i <= workers; i++ {
    wg.Add(1)
    go copyWorker(i, jobQueue, errorChan, &wg)
}

// Send jobs
for _, file := range files {
    jobQueue <- copyJob{src: file.src, dst: file.dst}
}

close(jobQueue)
wg.Wait()
```

**Streaming Progress Reports**:
```go
progressChan := make(chan int64, 10)  // Buffered to avoid blocking
go func() {
    defer close(progressChan)
    for {
        n, err := src.Read(buffer)
        if n > 0 {
            dst.Write(buffer[:n])
            progressChan <- int64(n)  // Non-blocking report
        }
        if err == io.EOF {
            break
        }
    }
}()
```

**Two-Phase Directory Copy**:
```go
// Phase 1: Create all directories (serial, deterministic)
filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
    if d.IsDir() {
        return os.MkdirAll(dstPath, 0755)
    }
    return nil
})

// Phase 2: Copy all files (parallel, safe from race conditions)
filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
    if !d.IsDir() {
        jobQueue <- copyJob{src: path, dst: dstPath}
    }
    return nil
})
```

## Error Handling Strategy

**Cascading errors**: Operations fail fast and propagate context
```go
// CopyFile errors preserve full context
if err != nil {
    log.Println("Error reading source file:", srcfile, err)
    return err
}

// Worker pool captures first error and stops
select {
case err := <-errorChan:
    return err
default:
    return nil
}
```

**Idempotent operations**: RemoveFile succeeds even if file doesn't exist
```go
if os.IsNotExist(err) {
    return nil  // Not an error - file already gone
}
```

**Partial failure cleanup**: Failed directory copies clean up
```go
if walkErr != nil {
    if removeErr := os.RemoveAll(dstDir); removeErr != nil {
        log.Println("failed to clean up destination:", removeErr)
    }
    return walkErr
}
```

## Testing Approach

Comprehensive TDD with Ginkgo/Gomega (50+ tests):

**Test Coverage**:
1. **File Operations** (13 tests) - Copy, move, delete, read, write with error cases
2. **Directory Operations** (8 tests) - Listing, creation, removal, traversal
3. **Metadata Operations** (6 tests) - Existence checks, sizes, MD5 hashing
4. **Error Handling** (8 tests) - Missing files, permission errors, partial failures
5. **Advanced Operations** (4 tests) - Progress tracking, worker pools
6. **Performance Tests** (3 tests) - Efficiency with large files and many files
7. **Integration Tests** (2 tests) - Complete workflows

**Test Execution**:
```bash
go test -v ./cmd/gstorage        # All tests with output
go test -race ./cmd/gstorage    # Race detector (verifies thread safety)
ginkgo -v ./cmd/gstorage        # Ginkgo-specific output format
```

## Running

```bash
# Run all tests
go test -v ./cmd/gstorage

# Run with race detector (verifies concurrency safety)
go test -race ./cmd/gstorage

# Run specific test suite
go test -v ./cmd/gstorage -run "CopyFile"
go test -v ./cmd/gstorage -run "WorkerPool"

# Run with Ginkgo for detailed output
ginkgo -v ./cmd/gstorage

# Run with coverage
go test -cover ./cmd/gstorage
```

## Example Usage

```go
package main

import (
    "fmt"
    "storage/cmd/gstorage"
)

func main() {
    // Simple file copy
    if err := gstorage.CopyFile("source.txt", "destination.txt"); err != nil {
        panic(err)
    }

    // Check file existence and size
    exists, _ := gstorage.FileExists("destination.txt")
    if exists {
        size, _ := gstorage.GetFileSize("destination.txt")
        fmt.Printf("File size: %d bytes\n", size)
    }

    // Calculate hash for verification
    hash, _ := gstorage.CalculateFileMD5("destination.txt")
    fmt.Printf("MD5: %s\n", hash)

    // Copy with progress tracking
    progressChan, _ := gstorage.CopyFileWithProgress("large.iso", "copy.iso", 64*1024)
    for bytes := range progressChan {
        fmt.Printf("Copied %d bytes\n", bytes)
    }

    // Parallel directory copy with worker pool
    if err := gstorage.WorkerPoolCopyDir("/source/dir", "/dest/dir", 4); err != nil {
        panic(err)
    }

    // Recursive directory operations
    gstorage.CreateDir("/path/to/deep/dir", true)
    gstorage.CopyDir("/src", "/dst")
    gstorage.RemoveDirAll("/cleanup/dir")
}
```

## Design Decisions

**Why separate functions instead of a File type?**
Simple operations don't need objects. Functional API is cleaner for one-off filesystem tasks.

**Why include MD5 hashing?**
File verification is essential in backup/sync scenarios. Included as a core operation rather than external dependency.

**Why worker pool for directory copy?**
Sequential copies across many files leave disk I/O idle. Parallel copies with bounded worker count saturate bandwidth efficiently without resource exhaustion.

**Why two-phase directory copy?**
Creating directories serially first ensures no race conditions where files get copied to non-existent destinations. Simpler and safer than lock-based coordination.

**Why progress reporting via channel?**
Allows callers to decide how to handle progress (log, update UI, etc.) without blocking the copy operation. Buffered channel prevents backpressure.

**Why separate RemoveFile (idempotent) vs RemoveDir?**
Different semantics: RemoveFile is idempotent (already gone = success), RemoveDir fails if directory not empty (safety check).

**Why log output in operations?**
Helps with debugging in production systems. Can be disabled by redirecting logs if needed.

## Real-World Applications

This pattern appears in:

- **Backup systems** (rsync, restic, duplicati) - efficient copying with verification
- **Container runtimes** (Docker, Containerd) - layer management, image copying
- **Deployment tools** (Kubernetes, Terraform) - file provisioning and cleanup
- **Package managers** (apt, brew, npm) - artifact copying and organization
- **File synchronization** (syncthing, Dropbox) - directory mirroring

## Skills Demonstrated

**Go Fundamentals**:
- File I/O with buffered operations
- Directory traversal with filepath.Walk and filepath.WalkDir
- Error handling with context preservation
- Goroutine-based concurrency
- Channel-based communication and worker pools

**Production Practices**:
- Idempotent operations (safe to retry)
- Proper error propagation and cleanup
- Resource management (file handles, goroutines)
- Performance optimization (streaming, parallelism)
- Observable operations (progress tracking, logging)

**Testing Rigor**:
- TDD with comprehensive test coverage
- Error path testing
- Race condition detection
- Performance benchmarking
- Integration test scenarios

## Background

Senior Platform Engineer (28 years) maintaining Go fundamentals between infrastructure projects. This repo represents deliberate practice - isolating filesystem operation patterns that underpin production infrastructure.

**Why share practice code?** Because seeing how senior engineers approach continuous learning matters. Not everything needs to be a finished product - sometimes exploring the fundamentals is the point.

## Development Notes

Built using AI-assisted pair programming (Claude) to explore modern development workflows while maintaining rigorous TDD practices and deep technical fundamentals.

## License

MIT
