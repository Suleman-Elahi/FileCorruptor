
# File Corruptor

A command-line tool written in Go to corrupt files on a public server, making them unusable after use. This acts as a shredder to prevent data theft by overwriting file contents with random data. Use with caution—this is a destructive tool!

## Features
- Corrupts individual files or directories
- Handles text files (`.txt`, `.csv`, `.msg`, etc.) by replacing contents with 32 bytes of random data
- For non-text files (audio, video, PDF, etc.):
  - Reduces to 1MB if originally larger
  - Keeps original size if ≤ 1MB, filled with random data
- Multi-threaded for faster processing
- Safety checks to prevent corruption of critical system files
- Option to override safety with a `--force` flag (use carefully)

## Installation & Usage

### Prerequisites
- [Go](https://go.dev/doc/install) (version 1.22 or later recommended, tested with 1.24.2)

### Build from Source
1. Clone or download this repository:
   ```bash
   git clone https://github.com/Suleman-Elahi/FileCorruptor
   cd file-corruptor
2.  Build the executable:
    
    ```bash
    go build -o fc .
    ```
    
    On Windows, use:    
    ```bash
    go build -o fc.exe .
    ```
    
3.  (Optional) Move fc/fc.exe to a directory in your PATH for global access.
    

### Usage:

Run the tool from the command line with the following options:

```bash
# Corrupt a single file
./fc --f filename

# Corrupt all files in a specific directory (top-level only)
./fc --dir foldername

# Corrupt all files in a directory and its subdirectories
./fc --dir foldername all

# Override safety checks (DANGEROUS, use with caution)
./fc --dir foldername --force
```
### Examples:
```bash
# Corrupt a single text file
./fc --f document.txt

# Corrupt files in 'test' directory (not subdirectories)
./fc --dir ./test

# Corrupt everything in 'test' and its subdirectories
./fc --dir ./test all
```

### Safety Features:

-   Protected Paths: Prevents corruption of critical system directories (e.g., C:\Windows, /usr, /System) unless --force is used.
    
-   Self-Protection: Skips corrupting the fc executable itself.
    
-   Directory Requirement: Requires a directory before all (e.g., --dir all fails).
    

### Warnings:

-   Irreversible: This tool permanently corrupts files with no recovery option. Test on non-critical data first!
    
-   Use Responsibly: Only use on files you own or have permission to modify. Avoid system directories unless you understand the risks.
    
-   Backup: Always back up important data before use.
    

### Development

-   Requirements: Standard Go library only (no external dependencies)
    
-   Contributing: Feel free to submit issues or pull requests on GitHub.
