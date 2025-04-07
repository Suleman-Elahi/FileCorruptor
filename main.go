package main

import (
    "crypto/rand"
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "strings"
    "sync"
)

var (
    textExtensions = map[string]bool{
        ".txt": true, ".csv": true, ".msg": true,
        ".log": true, ".ini": true, ".cfg": true,
    }

    protectedPaths = map[string]bool{
        `c:\windows`:           true,
        `c:\program files`:     true,
        `c:\program files (x86)`: true,
        `c:\programdata`:       true,
        `c:\users`:             true,
        `/system`:              true,
        `/library`:             true,
        `/users`:               true,
        `/applications`:        true,
        `/bin`:                 true,
        `/sbin`:                true,
        `/etc`:                 true,
        `/usr`:                 true,
        `/var`:                 true,
        `/lib`:                 true,
        `/lib64`:               true,
        `/boot`:                true,
        `/root`:                true,
        `/home`:                true,
    }
)

func isProtectedPath(path string) bool {
    absPath, err := filepath.Abs(filepath.Clean(path))
    if err != nil {
        return true
    }
    
    absPath = strings.ToLower(absPath)
    
    for protected := range protectedPaths {
        if runtime.GOOS == "windows" {
            protected = strings.ToLower(protected)
        }
        if absPath == protected || strings.HasPrefix(absPath, protected+string(os.PathSeparator)) {
            return true
        }
    }
    return false
}

func corruptTextFile(filePath string) error {
    file, err := os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

    randomBytes := make([]byte, 32)
    _, err = rand.Read(randomBytes)
    if err != nil {
        return err
    }

    _, err = file.Write(randomBytes)
    return err
}

func corruptLargeFile(filePath string) error {
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        return err
    }
    originalSize := fileInfo.Size()

    file, err := os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

    const oneMB = 1024 * 1024
    var bufferSize int64

    if originalSize > oneMB {
        bufferSize = oneMB
    } else if originalSize > 0 {
        bufferSize = originalSize
    } else {
        bufferSize = oneMB
    }

    randomBytes := make([]byte, bufferSize)
    _, err = rand.Read(randomBytes)
    if err != nil {
        return err
    }

    _, err = file.Write(randomBytes)
    return err
}

func processFile(filePath string, exePath string) error {
    absFilePath, _ := filepath.Abs(filePath)
    absExePath, _ := filepath.Abs(exePath)
    if absFilePath == absExePath {
        return nil
    }

    if isProtectedPath(filePath) {
        return fmt.Errorf("file is in a protected system directory")
    }

    ext := strings.ToLower(filepath.Ext(filePath))
    if textExtensions[ext] {
        return corruptTextFile(filePath)
    }
    return corruptLargeFile(filePath)
}

func main() {
    fileName := flag.String("f", "", "File to corrupt")
    dirName := flag.String("dir", "", "Directory to corrupt")
    force := flag.Bool("force", false, "Force corruption even in protected directories (use with caution)")
    flag.Parse()

    exePath, err := os.Executable()
    if err != nil {
        fmt.Printf("Error getting executable path: %v\n", err)
        os.Exit(1)
    }

    args := flag.Args()
    recursive := false
    if len(args) > 0 && args[0] == "all" {
        recursive = true
    }

    if *fileName == "" && *dirName == "" {
        fmt.Println("Error: No arguments specified")
        fmt.Println("Usage:")
        fmt.Println("  fc --f filename          : corrupt a specific file")
        fmt.Println("  fc --dir foldername      : corrupt files in specified directory only")
        fmt.Println("  fc --dir foldername all  : corrupt all files in directory and subdirs")
        fmt.Println("  -force                  : override safety checks (DANGEROUS)")
        os.Exit(1)
    }

    if *dirName == "all" || (*dirName == "" && recursive) {
        fmt.Println("Error: Must specify a directory before 'all' argument")
        fmt.Println("Usage: fc --dir foldername all")
        os.Exit(1)
    }

    var wg sync.WaitGroup
    errChan := make(chan error, 10)

    if *fileName != "" {
        if _, err := os.Stat(*fileName); os.IsNotExist(err) {
            fmt.Printf("Error: File %s does not exist\n", *fileName)
            os.Exit(1)
        }
        if !*force && isProtectedPath(*fileName) {
            fmt.Printf("Error: %s is in a protected system directory. Use --force to override (not recommended)\n", *fileName)
            os.Exit(1)
        }
        fmt.Printf("Corrupting file: %s\n", *fileName)
        wg.Add(1)
        go func() {
            defer wg.Done()
            if err := processFile(*fileName, exePath); err != nil {
                errChan <- fmt.Errorf("Error corrupting %s: %v", *fileName, err)
            }
        }()
    }

    if *dirName != "" {
        path := *dirName
        if _, err := os.Stat(path); os.IsNotExist(err) {
            fmt.Printf("Error: Directory %s does not exist\n", path)
            os.Exit(1)
        }

        if !*force && isProtectedPath(path) {
            fmt.Printf("Error: %s is a protected system directory. Use --force to override (not recommended)\n", path)
            os.Exit(1)
        }

        fmt.Printf("Corrupting directory: %s\n", path)

        if recursive {
            err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
                if err != nil {
                    return err
                }
                if !info.IsDir() {
                    if !*force && isProtectedPath(filePath) {
                        return fmt.Errorf("%s is in a protected system directory", filePath)
                    }
                    wg.Add(1)
                    go func(fp string) {
                        defer wg.Done()
                        if err := processFile(fp, exePath); err != nil {
                            errChan <- fmt.Errorf("Error corrupting %s: %v", fp, err)
                        }
                    }(filePath)
                }
                return nil
            })
        } else {
            dir, err := os.Open(path)
            if err != nil {
                fmt.Printf("Error opening directory: %v\n", err)
                os.Exit(1)
            }
            defer dir.Close()

            files, err := dir.Readdir(-1)
            if err != nil {
                fmt.Printf("Error reading directory: %v\n", err)
                os.Exit(1)
            }

            for _, file := range files {
                if !file.IsDir() {
                    fullPath := filepath.Join(path, file.Name())
                    if !*force && isProtectedPath(fullPath) {
                        errChan <- fmt.Errorf("%s is in a protected system directory", fullPath)
                        continue
                    }
                    wg.Add(1)
                    go func(fp string) {
                        defer wg.Done()
                        if err := processFile(fp, exePath); err != nil {
                            errChan <- fmt.Errorf("Error corrupting %s: %v", fp, err)
                        }
                    }(fullPath)
                }
            }
        }

        if err != nil {
            fmt.Printf("Error processing directory: %v\n", err)
            os.Exit(1)
        }
    }

    go func() {
        wg.Wait()
        close(errChan)
    }()

    hasErrors := false
    for err := range errChan {
        fmt.Println(err)
        hasErrors = true
    }

    if hasErrors {
        fmt.Println("Completed with errors")
        os.Exit(1)
    }
    fmt.Println("Corruption completed successfully")
}