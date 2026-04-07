#!/usr/bin/env pwsh

# Скрипт для сравнения вывода mygrep и оригинального grep
# Используется для демонстрации корректности распределенной обработки.

$testFile = "testdata.txt"
$port1 = 6001
$port2 = 6002

# 1. Собираем исполняемые файлы
go build -o coordinator.exe coordinator.go
go build -o worker.exe worker.go

# 2. Запускаем воркеры в фоновом режиме
Write-Host "Starting workers..." -ForegroundColor Cyan
$w1 = Start-Process .\worker.exe -ArgumentList "-port :$port1 -callback http://localhost:6000" -PassThru -WindowStyle Hidden
$w2 = Start-Process .\worker.exe -ArgumentList "-port :$port2 -callback http://localhost:6000" -PassThru -WindowStyle Hidden
Start-Sleep -s 2

function Compare-Grep {
    param($title, $mygrepArgs, $grepPattern, $grepFlags)
    
    Write-Host "`n--- Testing: $title ---" -ForegroundColor Yellow
    
    # Подготовка аргументов для нашего mygrep
    $baseArgs = @("--pattern", "$grepPattern", "--servers", "http://localhost:$port1,http://localhost:$port2", "--file", $testFile, "--quorum", "2")
    $finalArgs = $baseArgs + $mygrepArgs
    
    # Запуск нашего mygrep
    $rawOutput = & ".\coordinator.exe" $finalArgs 2>&1
    $myResult = $rawOutput | ForEach-Object { $_.ToString().Trim() } | Where-Object { 
        $_ -and 
        $_ -notmatch "^\[COORDINATOR\]" -and 
        $_ -notmatch "^\d{4}/\d{2}/\d{2}" -and 
        $_ -notmatch "^Usage of" -and
        $_ -notmatch "^--" -and
        $_ -notmatch "^unknown flag:" -and
        $_ -notmatch "Accepted|Posted|Received response"
    } | Sort-Object
    
    # Запуск системного grep или эмуляция через PowerShell
    if (Get-Command grep -ErrorAction SilentlyContinue) {
        $expectedResult = grep $grepFlags "$grepPattern" $testFile | Sort-Object
    } else {
        $isCaseSensitive = $true
        if ($grepFlags -contains "-i") { $isCaseSensitive = $false }
        
        if ($grepFlags -contains "-A") {
            $contextLines = [int]$grepFlags[$grepFlags.IndexOf("-A") + 1]
            $allLines = Get-Content $testFile
            $indices = 0..($allLines.Count - 1) | Where-Object { $allLines[$_] -match ( [regex]::Escape($grepPattern) ) }
            $resultIndices = @()
            foreach ($idx in $indices) {
                $start = $idx
                $end = [Math]::Min($idx + $contextLines, $allLines.Count - 1)
                $resultIndices += $start..$end
            }
            $expectedResult = $resultIndices | Select-Object -Unique | ForEach-Object { $allLines[$_] } | Sort-Object
        } elseif ($grepFlags -contains "-v") {
            $expectedResult = Get-Content $testFile | Where-Object { $_ -notmatch ( [regex]::Escape($grepPattern) ) } | Sort-Object
        } else {
            $expectedResult = Get-Content $testFile | Select-String -Pattern "$grepPattern" -CaseSensitive:$isCaseSensitive | ForEach-Object { $_.Line } | Sort-Object
        }
    }

    if ($null -eq $myResult) { $myResult = @() }
    if ($null -eq $expectedResult) { $expectedResult = @() }

    $diff = Compare-Object $myResult $expectedResult
    if ($null -eq $diff) {
        Write-Host "SUCCESS: Results match!" -ForegroundColor Green
    } else {
        Write-Host "FAILURE: Results differ!" -ForegroundColor Red
        $diff | Format-Table
    }
}

# Тест 1: Простой поиск (case-sensitive)
Compare-Grep "Simple Search" @("--chunksize", "5") "ERROR" @()

# Тест 2: Поиск без учета регистра
Compare-Grep "Ignore Case" @("--chunksize", "5", "--i") "error" @("-i")

# Тест 3: Инвертированный поиск
Compare-Grep "Invert Match" @("--chunksize", "5", "--v") "normal" @("-v")

# Тест 4: Поиск с контекстом (After)
Compare-Grep "Context After" @("--chunksize", "10", "--A", "1") "failure" @("-A", "1")

# Остановка воркеров
Write-Host "`nStopping workers..." -ForegroundColor Cyan
Stop-Process -Id $w1.Id -Force
Stop-Process -Id $w2.Id -Force

# Очистка
Remove-Item coordinator.exe, worker.exe
