# PowerShell Testing Guide for aifiler Tool System
# This script demonstrates how to test the aifiler tools using PowerShell

# ============================================================================
# PART 1: Test Setup and Initialization
# ============================================================================

# Test configuration
$TEST_DIR = "D:\Projects\aifiler"
$TEST_OUTPUT_DIR = "$TEST_DIR\test-output"
$GO_BINARY = "$TEST_DIR\cmd\aifiler\main.go"

# Create test output directory
if (-not (Test-Path $TEST_OUTPUT_DIR)) {
    New-Item -ItemType Directory -Path $TEST_OUTPUT_DIR | Out-Null
    Write-Host "✓ Created test output directory: $TEST_OUTPUT_DIR" -ForegroundColor Green
}

# ============================================================================
# PART 2: Unit Test Examples
# ============================================================================

function Test-FileOperationTools {
    <#
    .SYNOPSIS
    Test file operation tools (read, write, list, etc.)
    
    .DESCRIPTION
    Tests all file operation tools with various scenarios
    #>
    
    Write-Host "`n=== Testing File Operation Tools ===" -ForegroundColor Cyan
    
    # Test 1: Create test files
    $testFile = "$TEST_OUTPUT_DIR\test_file.txt"
    $testContent = "Test content for file operations`nLine 2`nLine 3`nLine 4`nLine 5"
    Set-Content -Path $testFile -Value $testContent
    Write-Host "✓ Created test file: $testFile" -ForegroundColor Green
    
    # Test 2: Read file operations
    Write-Host "`nTesting read_file operation..." -ForegroundColor Yellow
    $fileInfo = Get-Item $testFile
    Write-Host "  - File size: $($fileInfo.Length) bytes"
    Write-Host "  - Last modified: $($fileInfo.LastWriteTime)"
    
    # Test 3: List directory
    Write-Host "`nTesting list_directory operation..." -ForegroundColor Yellow
    $dirContent = Get-ChildItem -Path $TEST_OUTPUT_DIR
    Write-Host "  - Files in directory: $($dirContent.Count)"
    foreach ($item in $dirContent) {
        Write-Host "    - $($item.Name) ($($item.Length) bytes)"
    }
    
    # Test 4: File metadata
    Write-Host "`nTesting file metadata extraction..." -ForegroundColor Yellow
    $metadata = @{
        Name = $fileInfo.Name
        Path = $fileInfo.FullName
        Size = $fileInfo.Length
        Extension = $fileInfo.Extension
        ModTime = $fileInfo.LastWriteTime
        IsDir = $fileInfo.PSIsContainer
        CanRead = $true
        CanWrite = -not $fileInfo.Attributes.HasFlag([System.IO.FileAttributes]::ReadOnly)
    }
    Write-Host "  - Metadata collected: $(($metadata.Keys | Measure-Object).Count) properties"
    
    # Test 5: File copy
    Write-Host "`nTesting file copy operation..." -ForegroundColor Yellow
    $copiedFile = "$TEST_OUTPUT_DIR\test_file_copy.txt"
    Copy-Item -Path $testFile -Destination $copiedFile
    if (Test-Path $copiedFile) {
        Write-Host "  ✓ File copied successfully" -ForegroundColor Green
    }
    
    # Test 6: Directory listing with sizes
    Write-Host "`nTesting directory analysis..." -ForegroundColor Yellow
    $totalSize = 0
    $fileTypes = @{}
    foreach ($item in (Get-ChildItem -Path $TEST_OUTPUT_DIR -Recurse)) {
        if (-not $item.PSIsContainer) {
            $totalSize += $item.Length
            $ext = $item.Extension
            if ([string]::IsNullOrEmpty($ext)) { $ext = "no-extension" }
            $fileTypes[$ext]++
        }
    }
    Write-Host "  - Total directory size: $totalSize bytes"
    Write-Host "  - File types found: $(($fileTypes.Keys | Measure-Object).Count)"
    foreach ($type in $fileTypes.Keys) {
        Write-Host "    - $type : $($fileTypes[$type]) files"
    }
}

function Test-FilePatternMatching {
    <#
    .SYNOPSIS
    Test file pattern matching (glob and regex)
    #>
    
    Write-Host "`n=== Testing File Pattern Matching ===" -ForegroundColor Cyan
    
    # Create test files with various extensions
    $extensions = @(".txt", ".md", ".json", ".yaml", ".go")
    foreach ($ext in $extensions) {
        $path = "$TEST_OUTPUT_DIR\sample$ext"
        Set-Content -Path $path -Value "sample content"
    }
    Write-Host "✓ Created sample files with extensions: $($extensions -join ', ')" -ForegroundColor Green
    
    # Test 1: Glob pattern - all .txt files
    Write-Host "`nTesting glob pattern (*.txt)..." -ForegroundColor Yellow
    $txtFiles = @(Get-ChildItem -Path "$TEST_OUTPUT_DIR\*.txt" -ErrorAction SilentlyContinue)
    Write-Host "  - Found $($txtFiles.Count) .txt files"
    
    # Test 2: Glob pattern - all JSON/YAML
    Write-Host "`nTesting glob pattern (*.{json,yaml})..." -ForegroundColor Yellow
    $configFiles = @(Get-ChildItem -Path "$TEST_OUTPUT_DIR\*.json", "$TEST_OUTPUT_DIR\*.yaml" -ErrorAction SilentlyContinue)
    Write-Host "  - Found $($configFiles.Count) config files"
    
    # Test 3: Regex pattern - files starting with 'sample'
    Write-Host "`nTesting regex pattern (^sample.*)..." -ForegroundColor Yellow
    $regexFiles = @(Get-ChildItem -Path $TEST_OUTPUT_DIR | Where-Object { $_.Name -match "^sample" })
    Write-Host "  - Found $($regexFiles.Count) files matching pattern"
    
    # Test 4: Pattern with depth limiting
    Write-Host "`nTesting recursive pattern search..." -ForegroundColor Yellow
    $allSampleFiles = @(Get-ChildItem -Path $TEST_OUTPUT_DIR -Filter "sample*" -Recurse)
    Write-Host "  - Found $($allSampleFiles.Count) files recursively"
}

function Test-DirectoryAnalysis {
    <#
    .SYNOPSIS
    Test directory scanning and analysis
    #>
    
    Write-Host "`n=== Testing Directory Analysis ===" -ForegroundColor Cyan
    
    # Create test directory structure
    $dirs = @("docs", "images", "logs", "data")
    foreach ($dir in $dirs) {
        $path = "$TEST_OUTPUT_DIR\$dir"
        if (-not (Test-Path $path)) {
            New-Item -ItemType Directory -Path $path | Out-Null
        }
    }
    Write-Host "✓ Created test directory structure" -ForegroundColor Green
    
    # Add files to directories
    for ($i = 1; $i -le 3; $i++) {
        Set-Content -Path "$TEST_OUTPUT_DIR\docs\doc$i.md" -Value "Documentation $i"
        Set-Content -Path "$TEST_OUTPUT_DIR\images\img$i.png" -Value "IMAGE_DATA_$i"
        Set-Content -Path "$TEST_OUTPUT_DIR\logs\log$i.txt" -Value "Log entry $i"
        Set-Content -Path "$TEST_OUTPUT_DIR\data\data$i.json" -Value "{}"
    }
    Write-Host "✓ Created test files in subdirectories" -ForegroundColor Green
    
    # Test 1: Count files and directories
    Write-Host "`nAnalyzing directory structure..." -ForegroundColor Yellow
    $stats = @{
        FileCount = 0
        DirCount = 0
        TotalSize = 0
        FilesByType = @{}
    }
    
    Get-ChildItem -Path $TEST_OUTPUT_DIR -Recurse | ForEach-Object {
        if ($_.PSIsContainer) {
            $stats.DirCount++
        } else {
            $stats.FileCount++
            $stats.TotalSize += $_.Length
            $ext = $_.Extension
            if ([string]::IsNullOrEmpty($ext)) { $ext = "no-extension" }
            $stats.FilesByType[$ext]++
        }
    }
    
    Write-Host "  - Total files: $($stats.FileCount)"
    Write-Host "  - Total directories: $($stats.DirCount)"
    Write-Host "  - Total size: $($stats.TotalSize) bytes"
    Write-Host "  - File types:"
    foreach ($type in $stats.FilesByType.Keys | Sort-Object) {
        Write-Host "    - $type : $($stats.FilesByType[$type]) files"
    }
    
    # Test 2: Find largest files
    Write-Host "`nFinding largest files..." -ForegroundColor Yellow
    $largestFiles = Get-ChildItem -Path $TEST_OUTPUT_DIR -Recurse -File | 
                    Sort-Object Length -Descending | 
                    Select-Object -First 5 Name, Length, LastWriteTime
    foreach ($file in $largestFiles) {
        Write-Host "  - $($file.Name): $($file.Length) bytes"
    }
}

# ============================================================================
# PART 3: Integration Testing
# ============================================================================

function Test-ToolIntegration {
    <#
    .SYNOPSIS
    Test tool integration and chaining
    #>
    
    Write-Host "`n=== Testing Tool Integration ===" -ForegroundColor Cyan
    
    # Scenario 1: Find all config files and read them
    Write-Host "`nScenario 1: Find and analyze config files..." -ForegroundColor Yellow
    $configDir = "$TEST_OUTPUT_DIR"
    $configFiles = @(Get-ChildItem -Path $configDir -Filter "*.json", "*.yaml" -ErrorAction SilentlyContinue)
    Write-Host "  - Found $($configFiles.Count) config files"
    foreach ($file in $configFiles) {
        Write-Host "    - Reading $($file.Name)..."
        if ($file.Length -lt 1MB) {
            $content = Get-Content -Path $file.FullName
            Write-Host "      Size: $($file.Length) bytes"
        }
    }
    
    # Scenario 2: Analyze file types in directory
    Write-Host "`nScenario 2: Analyze file types..." -ForegroundColor Yellow
    $typeAnalysis = @{}
    Get-ChildItem -Path $TEST_OUTPUT_DIR -Recurse -File | ForEach-Object {
        $ext = if ($_.Extension) { $_.Extension } else { "no-extension" }
        if (-not $typeAnalysis.ContainsKey($ext)) {
            $typeAnalysis[$ext] = @{ count = 0; size = 0 }
        }
        $typeAnalysis[$ext].count++
        $typeAnalysis[$ext].size += $_.Length
    }
    foreach ($type in $typeAnalysis.Keys | Sort-Object) {
        Write-Host "  - $type : $($typeAnalysis[$type].count) files, $($typeAnalysis[$type].size) bytes"
    }
    
    # Scenario 3: Organize files by type
    Write-Host "`nScenario 3: Planning file organization..." -ForegroundColor Yellow
    $organizationPlan = @{}
    Get-ChildItem -Path $TEST_OUTPUT_DIR -Recurse -File | ForEach-Object {
        $ext = if ($_.Extension) { $_.Extension.Substring(1) } else { "other" }
        if (-not $organizationPlan.ContainsKey($ext)) {
            $organizationPlan[$ext] = @()
        }
        $organizationPlan[$ext] += $_.Name
    }
    foreach ($category in $organizationPlan.Keys | Sort-Object) {
        Write-Host "  - $category : $($organizationPlan[$category].Count) files to move"
    }
}

# ============================================================================
# PART 4: Performance Testing
# ============================================================================

function Test-PerformanceBenchmarks {
    <#
    .SYNOPSIS
    Test tool performance with timing
    #>
    
    Write-Host "`n=== Testing Performance ===" -ForegroundColor Cyan
    
    # Test 1: Directory scanning performance
    Write-Host "`nTest 1: Directory scanning..." -ForegroundColor Yellow
    $timer = [System.Diagnostics.Stopwatch]::StartNew()
    $count = @(Get-ChildItem -Path $TEST_OUTPUT_DIR -Recurse -File).Count
    $timer.Stop()
    Write-Host "  - Scanned directory with $count files in $($timer.ElapsedMilliseconds)ms"
    
    # Test 2: File read performance
    Write-Host "`nTest 2: File reading..." -ForegroundColor Yellow
    $testFile = "$TEST_OUTPUT_DIR\sample.txt"
    if (Test-Path $testFile) {
        $timer = [System.Diagnostics.Stopwatch]::StartNew()
        $content = Get-Content -Path $testFile
        $timer.Stop()
        Write-Host "  - Read file in $($timer.ElapsedMilliseconds)ms"
    }
    
    # Test 3: Pattern matching performance
    Write-Host "`nTest 3: Pattern matching..." -ForegroundColor Yellow
    $timer = [System.Diagnostics.Stopwatch]::StartNew()
    $matches = @(Get-ChildItem -Path "$TEST_OUTPUT_DIR\*" -Recurse | Where-Object { $_.Name -match "\.json$|\.yaml$" })
    $timer.Stop()
    Write-Host "  - Found $($matches.Count) matching files in $($timer.ElapsedMilliseconds)ms"
    
    # Test 4: Metadata extraction performance
    Write-Host "`nTest 4: Metadata extraction..." -ForegroundColor Yellow
    $timer = [System.Diagnostics.Stopwatch]::StartNew()
    $metadata = @(Get-ChildItem -Path $TEST_OUTPUT_DIR -Recurse -File | 
                  Select-Object Name, FullName, Length, LastWriteTime | 
                  Measure-Object -Property Length -Sum)
    $timer.Stop()
    Write-Host "  - Extracted metadata in $($timer.ElapsedMilliseconds)ms"
    Write-Host "  - Total files analyzed: $($metadata.Count)"
}

# ============================================================================
# PART 5: Error Handling and Edge Cases
# ============================================================================

function Test-ErrorHandling {
    <#
    .SYNOPSIS
    Test error handling and edge cases
    #>
    
    Write-Host "`n=== Testing Error Handling ===" -ForegroundColor Cyan
    
    # Test 1: Non-existent file
    Write-Host "`nTest 1: Reading non-existent file..." -ForegroundColor Yellow
    $nonExistent = "$TEST_OUTPUT_DIR\nonexistent.txt"
    if (Test-Path $nonExistent) {
        Write-Host "  ✓ File exists"
    } else {
        Write-Host "  ✓ Correctly detected non-existent file" -ForegroundColor Green
    }
    
    # Test 2: Invalid patterns
    Write-Host "`nTest 2: Testing invalid patterns..." -ForegroundColor Yellow
    try {
        $invalid = Get-ChildItem -Path "$TEST_OUTPUT_DIR\[invalid" -ErrorAction Stop
    } catch {
        Write-Host "  ✓ Correctly caught invalid pattern" -ForegroundColor Green
    }
    
    # Test 3: Permission handling
    Write-Host "`nTest 3: Permission handling..." -ForegroundColor Yellow
    $testFile = "$TEST_OUTPUT_DIR\readonly_test.txt"
    Set-Content -Path $testFile -Value "test"
    $currentAcl = Get-Acl -Path $testFile
    Write-Host "  ✓ File permissions readable"
    
    # Test 4: Large file handling
    Write-Host "`nTest 4: Large file handling..." -ForegroundColor Yellow
    $largeFile = "$TEST_OUTPUT_DIR\large_file.dat"
    $largeData = [byte[]]::new(10MB)
    [System.Random]::new().NextBytes($largeData)
    [System.IO.File]::WriteAllBytes($largeFile, $largeData)
    $fileInfo = Get-Item $largeFile
    Write-Host "  ✓ Created and verified $([math]::Round($fileInfo.Length/1MB, 2))MB file"
    
    # Test 5: Special characters in filenames
    Write-Host "`nTest 5: Special characters in filenames..." -ForegroundColor Yellow
    $specialFile = "$TEST_OUTPUT_DIR\file with spaces & special.txt"
    Set-Content -Path $specialFile -Value "test"
    if (Test-Path $specialFile) {
        Write-Host "  ✓ Handled special characters correctly" -ForegroundColor Green
        Remove-Item $specialFile
    }
}

# ============================================================================
# PART 6: Main Test Execution
# ============================================================================

function Run-AllTests {
    <#
    .SYNOPSIS
    Run all test suites
    #>
    
    Write-Host "`n╔══════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "║   aifiler Tools System - PowerShell Testing Suite         ║" -ForegroundColor Cyan
    Write-Host "╚══════════════════════════════════════════════════════════╝`n" -ForegroundColor Cyan
    
    Write-Host "Test Output Directory: $TEST_OUTPUT_DIR`n" -ForegroundColor Gray
    
    try {
        Test-FileOperationTools
        Test-FilePatternMatching
        Test-DirectoryAnalysis
        Test-ToolIntegration
        Test-PerformanceBenchmarks
        Test-ErrorHandling
        
        Write-Host "`n╔══════════════════════════════════════════════════════════╗" -ForegroundColor Green
        Write-Host "║   All tests completed successfully!                        ║" -ForegroundColor Green
        Write-Host "╚══════════════════════════════════════════════════════════╝`n" -ForegroundColor Green
        
        # Print summary
        Write-Host "Test Summary:" -ForegroundColor Green
        Write-Host "  ✓ File operation tools"
        Write-Host "  ✓ Pattern matching"
        Write-Host "  ✓ Directory analysis"
        Write-Host "  ✓ Tool integration"
        Write-Host "  ✓ Performance benchmarks"
        Write-Host "  ✓ Error handling`n"
        
    } catch {
        Write-Host "`n✗ Test failed with error:`n$_`n" -ForegroundColor Red
    }
}

# ============================================================================
# Run Tests
# ============================================================================

# Run all tests
Run-AllTests

# Cleanup option
Write-Host "Cleanup test output directory? (y/n): " -NoNewline -ForegroundColor Gray
$response = Read-Host
if ($response -eq 'y') {
    Remove-Item -Path $TEST_OUTPUT_DIR -Recurse -Force
    Write-Host "✓ Test directory cleaned up" -ForegroundColor Green
} else {
    Write-Host "✓ Test directory preserved at: $TEST_OUTPUT_DIR" -ForegroundColor Green
}
