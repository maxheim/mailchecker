package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Config structure for JSON config file
type Config struct {
	Folders []string `json:"folders"`
}

// FolderResult stores the results for a single folder
type FolderResult struct {
	FolderPath     string
	DateCountMap   map[string]int
	FileCountMap   map[string]int
	DateHourlyData map[string]map[int]int // date -> hour -> count
	TotalCount     int
	Error          error
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	var folderPaths []string
	verbose := false

	// Parse command line arguments
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--verbose" {
			verbose = true
		} else if arg == "--config" {
			if i+1 >= len(os.Args) {
				fmt.Println("Error: --config flag requires a file path")
				os.Exit(1)
			}
			configPath := os.Args[i+1]
			paths, err := loadConfigFile(configPath)
			if err != nil {
				fmt.Printf("Error loading config file: %v\n", err)
				os.Exit(1)
			}
			folderPaths = append(folderPaths, paths...)
			i++ // Skip next argument (config file path)
		} else if !strings.HasPrefix(arg, "--") {
			// It's a folder path
			folderPaths = append(folderPaths, arg)
		}
	}

	if len(folderPaths) == 0 {
		fmt.Println("Error: No folder paths provided")
		printUsage()
		os.Exit(1)
	}

	fmt.Printf("Analyzing %d folder(s)...\n", len(folderPaths))

	// Process folders concurrently
	results := processFoldersConcurrently(folderPaths)

	// Aggregate results
	aggregateDateCountMap := make(map[string]int)
	totalEntriesAcrossAllFolders := 0
	successfulFolders := 0

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("RESULTS BY FOLDER")
	fmt.Println(strings.Repeat("=", 80))

	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("\n[ERROR] Folder: %s\n", result.FolderPath)
			fmt.Printf("  Error: %v\n", result.Error)
			continue
		}

		successfulFolders++
		fmt.Printf("\n[SUCCESS] Folder: %s\n", result.FolderPath)

		// Show per-file counts if verbose mode is enabled
		if verbose && len(result.FileCountMap) > 0 {
			fmt.Println("  Files:")
			for fileName, count := range result.FileCountMap {
				fmt.Printf("    - %s: %d entries\n", fileName, count)
			}
		}

		// Show per-day statistics with average emails per hour if verbose mode is enabled
		if verbose && len(result.DateCountMap) > 0 {
			fmt.Println("  Per-Day Statistics:")
			for date, count := range result.DateCountMap {
				// Calculate average emails per hour for this date
				avgPerHour := calculateAveragePerHour(result.DateHourlyData[date], count)
				fmt.Printf("    - %s: %d entries (avg %.2f emails/hour)\n", date, count, avgPerHour)
			}
		}

		fmt.Printf("  Total '2FA - Email' entries: %d\n", result.TotalCount)
		totalEntriesAcrossAllFolders += result.TotalCount

		// Aggregate dates
		for date, count := range result.DateCountMap {
			aggregateDateCountMap[date] += count
		}
	}

	// Print aggregate summary
	if len(aggregateDateCountMap) == 0 {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("No entries with '2FA - Email' found in any log files.")
		return
	}

	distinctDays := len(aggregateDateCountMap)
	average := float64(totalEntriesAcrossAllFolders) / float64(distinctDays)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("AGGREGATE RESULTS (ALL FOLDERS)")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("\n2FA - Email Entries by Date:")
	for date, count := range aggregateDateCountMap {
		fmt.Printf("  %s: %d entries\n", date, count)
	}

	fmt.Println("\nSummary:")
	fmt.Printf("  Total folders processed: %d\n", len(folderPaths))
	fmt.Printf("  Successful folders: %d\n", successfulFolders)
	fmt.Printf("  Total entries with '2FA - Email': %d\n", totalEntriesAcrossAllFolders)
	fmt.Printf("  Total distinct days: %d\n", distinctDays)
	fmt.Printf("  Average entries per day: %.2f\n", average)
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  analyze_logs [options] <folder_path1> [folder_path2] ...")
	fmt.Println("  analyze_logs [options] --config <config_file>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --verbose       Show detailed per-file statistics")
	fmt.Println("  --config <file> Load folder paths from a JSON config file")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  analyze_logs C:\\Logs\\Folder1")
	fmt.Println("  analyze_logs C:\\Logs\\Folder1 D:\\Logs\\Folder2 --verbose")
	fmt.Println("  analyze_logs --config config.json --verbose")
	fmt.Println("  analyze_logs \\\\\\server1\\share\\logs \\\\\\server2\\share\\logs")
}

func loadConfigFile(configPath string) ([]string, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if len(config.Folders) == 0 {
		return nil, fmt.Errorf("no folders specified in config file")
	}

	return config.Folders, nil
}

func processFoldersConcurrently(folderPaths []string) []FolderResult {
	var wg sync.WaitGroup
	results := make([]FolderResult, len(folderPaths))

	for i, folderPath := range folderPaths {
		wg.Add(1)
		go func(index int, path string) {
			defer wg.Done()
			results[index] = processFolder(path)
		}(i, folderPath)
	}

	wg.Wait()
	return results
}

func calculateAveragePerHour(hourlyData map[int]int, totalCount int) float64 {
	if len(hourlyData) == 0 {
		return 0.0
	}

	// Find min and max hour to determine the time span
	minHour, maxHour := 23, 0
	for hour := range hourlyData {
		if hour < minHour {
			minHour = hour
		}
		if hour > maxHour {
			maxHour = hour
		}
	}

	// Calculate hours span (inclusive)
	hoursSpan := maxHour - minHour + 1
	if hoursSpan <= 0 {
		hoursSpan = 1
	}

	return float64(totalCount) / float64(hoursSpan)
}

func processFolder(folderPath string) FolderResult {
	result := FolderResult{
		FolderPath:     folderPath,
		DateCountMap:   make(map[string]int),
		FileCountMap:   make(map[string]int),
		DateHourlyData: make(map[string]map[int]int),
	}

	// Read all .txt files in the folder
	files, err := filepath.Glob(filepath.Join(folderPath, "*.txt"))
	if err != nil {
		result.Error = fmt.Errorf("error reading folder: %w", err)
		return result
	}

	if len(files) == 0 {
		result.Error = fmt.Errorf("no .txt files found in folder")
		return result
	}

	// Process each file
	for _, filePath := range files {
		file, err := os.Open(filePath)
		if err != nil {
			// Log error but continue with other files
			fmt.Printf("Warning: Error opening file %s: %v\n", filePath, err)
			continue
		}

		fileName := filepath.Base(filePath)
		fileCount := 0

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()

			// Check if line contains "2FA - Email"
			if strings.Contains(line, "2FA - Email") {
				// Extract the date and time from the line (format: YYYY-MM-DD HH:MM:SS)
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					dateStr := parts[0]
					timeStr := parts[1]

					// Parse date to ensure it's valid
					_, err := time.Parse("2006-01-02", dateStr)
					if err == nil {
						result.DateCountMap[dateStr]++
						fileCount++
						result.TotalCount++

						// Extract hour from time string (HH:MM:SS)
						timeParts := strings.Split(timeStr, ":")
						if len(timeParts) >= 1 {
							var hour int
							_, err := fmt.Sscanf(timeParts[0], "%d", &hour)
							if err == nil && hour >= 0 && hour <= 23 {
								// Initialize map for this date if needed
								if result.DateHourlyData[dateStr] == nil {
									result.DateHourlyData[dateStr] = make(map[int]int)
								}
								result.DateHourlyData[dateStr][hour]++
							}
						}
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Warning: Error reading file %s: %v\n", filePath, err)
		}

		result.FileCountMap[fileName] = fileCount
		file.Close()
	}

	return result
}
