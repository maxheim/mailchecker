# 2FA Log Analyzer

A Go-based command-line tool for analyzing 2FA authentication logs across multiple folders, including network locations. This tool efficiently processes log files to count "2FA - Email" entries, groups them by date, and calculates daily averages.

## Features

- **Multiple Folder Support**: Analyze logs from multiple folders in a single run
- **Network Path Support**: Works with Windows UNC paths (\\\\server\\share)
- **Concurrent Processing**: Processes multiple folders in parallel for optimal performance
- **Config File Support**: Maintain a list of folders in a JSON config file
- **Flexible Input**: Use command-line arguments, config files, or both
- **Verbose Mode**: Get detailed per-file statistics
- **Aggregated Results**: View combined statistics across all folders
- **Error Handling**: Continues processing even if individual folders fail

## Prerequisites

- **Go**: Version 1.16 or higher
- **Network Access**: Proper permissions to access network shares (if using UNC paths)
- **Windows**: For network path support (UNC paths like \\\\server\\share)

## Installation

### Option 1: Run from source
```bash
go run analyze_logs.go [options] <folder_paths...>
```

### Option 2: Build executable
```bash
# Build for Windows
go build -o analyze_logs.exe analyze_logs.go

# Run the executable
analyze_logs.exe [options] <folder_paths...>
```

### Option 3: Cross-compile for Windows (from Mac/Linux)
```bash
GOOS=windows GOARCH=amd64 go build -o analyze_logs.exe analyze_logs.go
```

## Usage

### Basic Syntax

```bash
analyze_logs [options] <folder_path1> [folder_path2] ...
analyze_logs [options] --config <config_file>
```

### Options

- `--verbose` : Show detailed per-file statistics
- `--config <file>` : Load folder paths from a JSON config file

### Examples

#### Single Folder
```bash
# Windows local path
go run analyze_logs.go C:\Logs\Production

# After building
analyze_logs.exe C:\Logs\Production
```

#### Multiple Folders (Command Line)
```bash
# Multiple local folders
go run analyze_logs.go C:\Logs\Folder1 D:\Logs\Folder2

# Network paths
go run analyze_logs.go \\server1\logs \\server2\logs

# Mixed local and network paths
go run analyze_logs.go C:\Logs\Local \\server\share\logs
```

#### Using Config File
```bash
# Basic usage
go run analyze_logs.go --config config.json

# With verbose output
go run analyze_logs.go --config config.json --verbose
```

#### Combining Config File and Command Line
```bash
# Analyze folders from config plus additional folders
go run analyze_logs.go --config config.json C:\Additional\Logs
```

#### Verbose Mode
```bash
# Show per-file statistics
go run analyze_logs.go --config config.json --verbose
```

## Configuration File

The config file is a JSON file that contains a list of folder paths to analyze.

### Format

```json
{
  "folders": [
    "C:\\Logs\\Production\\Server1",
    "C:\\Logs\\Production\\Server2",
    "\\\\FILESERVER01\\SharedLogs\\Application",
    "\\\\FILESERVER02\\Logs\\WebApp",
    "\\\\192.168.1.100\\LogShare\\2FA",
    "D:\\LocalLogs\\Backup"
  ]
}
```

### Path Format Notes

- **Windows Local Paths**: Use double backslashes (`C:\\Logs\\Folder`)
- **UNC Network Paths**: Use four backslashes for the server (`\\\\server\\share`)
- **IP Address Paths**: `\\\\192.168.1.100\\share\\folder`

### Creating Your Config File

1. Copy the provided `config.json` template
2. Replace the example paths with your actual folder locations
3. Ensure proper escaping of backslashes in JSON
4. Save the file in the same directory as the script

## Log File Format

The script expects log files with the following characteristics:

- **File Extension**: `.txt`
- **Date Format**: Each line must start with `YYYY-MM-DD HH:MM:SS`
- **Search String**: Lines containing `2FA - Email` will be counted
- **Example Line**:
  ```
  2024-01-15 14:23:45 [INFO] User authentication process: 2FA - Email - Session ID: 12345 - Status: Success
  ```

## Output

### Standard Output

```
Analyzing 3 folder(s)...

================================================================================
RESULTS BY FOLDER
================================================================================

[SUCCESS] Folder: C:\Logs\Folder1
  Total '2FA - Email' entries: 314

[SUCCESS] Folder: \\server\logs
  Total '2FA - Email' entries: 292

[ERROR] Folder: C:\Logs\Folder3
  Error: no .txt files found in folder

================================================================================
AGGREGATE RESULTS (ALL FOLDERS)
================================================================================

2FA - Email Entries by Date:
  2024-01-15: 314 entries
  2024-02-20: 292 entries

Summary:
  Total folders processed: 3
  Successful folders: 2
  Total entries with '2FA - Email': 606
  Total distinct days: 2
  Average entries per day: 303.00
```

### Verbose Output

When using `--verbose`, additional per-file details are shown:

```
[SUCCESS] Folder: C:\Logs\Folder1
  Files:
    - log_2024-01-15.txt: 314 entries
    - log_2024-01-20.txt: 287 entries
  Total '2FA - Email' entries: 601
```

## Network Paths on Windows

### UNC Path Format
```
\\servername\sharename\path\to\logs
\\192.168.1.100\LogShare\Application
```

### Accessing Network Shares

Before running the script, ensure you have:

1. **Network Connectivity**: Can ping the server
2. **Proper Permissions**: Read access to the network share
3. **Mounted Shares** (optional): Map network drive for easier access

### Mapping Network Drive (Optional)

```cmd
# Map network drive
net use Z: \\server\share /persistent:yes

# Use mapped drive in script
analyze_logs.exe Z:\logs
```

### Troubleshooting Network Access

If you encounter "Access Denied" errors:

```cmd
# Check current network connections
net use

# Connect with credentials
net use \\server\share /user:DOMAIN\username password

# Test access
dir \\server\share\logs
```

## Performance Considerations

### Concurrent Processing

The script processes multiple folders concurrently using goroutines:
- Each folder is processed in parallel
- Reduces total execution time, especially with network paths
- Network latency is minimized through parallel I/O

### Network Performance Tips

1. **Use Config File**: Faster than typing long UNC paths repeatedly
2. **Map Drives**: Sometimes faster than direct UNC paths
3. **Local Copies**: For repeated analysis, consider copying logs locally first
4. **Batch Processing**: Process all folders at once rather than multiple runs

## Error Handling

The script continues processing even if individual folders fail:

- **Missing Folder**: Reports error, continues with other folders
- **No .txt Files**: Reports error, continues processing
- **Permission Denied**: Reports error, continues processing
- **Network Timeout**: Reports error, continues processing

## Troubleshooting

### Common Issues

#### "No .txt files found in folder"
- Verify the folder path is correct
- Check that .txt files exist in the folder
- Ensure you have read permissions

#### "Error reading folder"
- Check folder path syntax (Windows: use `\` or escaped `\\`)
- Verify folder exists
- Check network connectivity (for UNC paths)

#### "Access Denied"
- Verify you have read permissions
- Try accessing the folder in File Explorer first
- Use `net use` to authenticate if needed

#### Config File Errors
- Ensure JSON is valid (use a JSON validator)
- Check backslash escaping (use `\\` for Windows paths)
- Verify the config file path is correct

### Testing Your Setup

```bash
# Test with local folder first
go run analyze_logs.go Examples

# Test with single network path
go run analyze_logs.go \\server\share\logs

# Test config file parsing
go run analyze_logs.go --config config.json
```

## Building for Production

### Build Optimized Executable

```bash
# Windows executable with optimizations
go build -ldflags="-s -w" -o analyze_logs.exe analyze_logs.go

# This produces a smaller executable by:
# -s: Omitting symbol table
# -w: Omitting DWARF debug info
```

### Deployment

1. Build the executable: `go build -o analyze_logs.exe analyze_logs.go`
2. Copy `analyze_logs.exe` to your target Windows machine
3. Create and configure `config.json` with your network paths
4. Run from Command Prompt or PowerShell

## Advanced Usage

### Scheduled Tasks (Windows)

Create a scheduled task to run analysis periodically:

```cmd
# Create batch file: run_analysis.bat
@echo off
cd C:\Path\To\Script
analyze_logs.exe --config config.json > output_%date:~-4,4%%date:~-10,2%%date:~-7,2%.txt
```

Schedule with Task Scheduler:
1. Open Task Scheduler
2. Create Basic Task
3. Set trigger (e.g., daily at 9 AM)
4. Action: Start a program
5. Program: `C:\Path\To\run_analysis.bat`

### Output Redirection

```bash
# Save output to file
analyze_logs.exe --config config.json > report.txt

# Append to existing file
analyze_logs.exe --config config.json >> monthly_report.txt

# Save with timestamp
analyze_logs.exe --config config.json > report_%date%.txt
```

## Support & Contributing

### Reporting Issues
- Provide the full error message
- Include your Go version: `go version`
- Describe your environment (OS, network setup)
- Include a sample log file (with sensitive data removed)

### Feature Requests
- Describe the use case
- Explain expected behavior
- Provide example scenarios

## License

This tool is provided as-is for internal use. Modify as needed for your requirements.

## Changelog

### Version 2.0
- Added multiple folder support
- Added concurrent processing
- Added config file support
- Added network path (UNC) support
- Added verbose mode
- Improved error handling
- Added aggregate results across folders

### Version 1.0
- Initial release
- Single folder analysis
- Basic date grouping and averaging
