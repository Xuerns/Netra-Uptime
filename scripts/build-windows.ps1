$ErrorActionPreference = "Stop"

$RepoRoot = Split-Path -Parent $PSScriptRoot
$ClientDir = Join-Path $RepoRoot "client"
$ClientBuildDir = Join-Path $ClientDir "build"
$ServerDir = Join-Path $RepoRoot "server"
$EmbedBuildDir = Join-Path $ServerDir "internal\static\build"
$OutputExe = Join-Path $RepoRoot "netra-uptime.exe"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host " Netra Uptime - Windows Standalone Binary Build Pipeline" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# Step 1: Build SvelteKit static frontend
Write-Host "[1/4] Building SvelteKit static frontend..." -ForegroundColor Yellow
pnpm --dir "$ClientDir" build
if ($LASTEXITCODE -ne 0) {
    throw "Frontend build failed with exit code $LASTEXITCODE"
}

try {
    # Step 2: Sync compiled frontend into Go embed directory
    Write-Host "[2/4] Syncing frontend build assets into server/internal/static/build..." -ForegroundColor Yellow
    if (Test-Path "$EmbedBuildDir") {
        Get-ChildItem -Path "$EmbedBuildDir" -Force | Remove-Item -Recurse -Force
    } else {
        New-Item -ItemType Directory -Path "$EmbedBuildDir" -Force | Out-Null
    }
    Copy-Item -Path "$ClientBuildDir\*" -Destination "$EmbedBuildDir" -Recurse -Force

    # Step 3: Compile Pure-Go Windows executable (CGO_ENABLED=0)
    Write-Host "[3/4] Compiling Pure-Go standalone binary (CGO_ENABLED=0, windows/amd64)..." -ForegroundColor Yellow
    Push-Location "$ServerDir"
    try {
        $env:CGO_ENABLED = "0"
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"
        go build -trimpath -ldflags="-s -w" -o "$OutputExe" ./cmd/netra-uptime
        if ($LASTEXITCODE -ne 0) {
            throw "Go build failed with exit code $LASTEXITCODE"
        }
    } finally {
        Pop-Location
    }
} finally {
    # Step 4: Clean up embedded build assets and restore placeholder index.html so Git stays clean
    Write-Host "[4/4] Cleaning temporary embed assets and restoring placeholder..." -ForegroundColor Yellow
    Get-ChildItem -Path "$EmbedBuildDir" -Force -ErrorAction SilentlyContinue | Remove-Item -Recurse -Force
    Push-Location "$RepoRoot"
    try {
        git checkout -- "server/internal/static/build/index.html" 2>$null
    } finally {
        Pop-Location
    }
}

if (Test-Path "$OutputExe") {
    $FileInfo = Get-Item "$OutputExe"
    $SizeMB = [math]::Round($FileInfo.Length / 1MB, 2)
    Write-Host "==========================================================" -ForegroundColor Green
    Write-Host " Build Successful!" -ForegroundColor Green
    Write-Host " Binary : $OutputExe" -ForegroundColor Green
    Write-Host " Size   : $SizeMB MB ($($FileInfo.Length) bytes)" -ForegroundColor Green
    Write-Host "==========================================================" -ForegroundColor Green
} else {
    throw "Expected output binary not found at $OutputExe"
}
