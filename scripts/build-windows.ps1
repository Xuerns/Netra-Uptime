param(
    [ValidateSet("desktop", "server")]
    [string]$Target = "desktop",
    [switch]$CreateZip
)

$ErrorActionPreference = "Stop"

$RepoRoot = Split-Path -Parent $PSScriptRoot
$ClientDir = Join-Path $RepoRoot "client"
$ClientBuildDir = Join-Path $ClientDir "build"
$ServerDir = Join-Path $RepoRoot "server"
$EmbedBuildDir = Join-Path $ServerDir "internal\static\build"
$OutputExe = Join-Path $RepoRoot "netra-uptime.exe"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host " Netra Uptime - Windows Standalone Binary Build Pipeline" -ForegroundColor Cyan
Write-Host " Target: $Target" -ForegroundColor Cyan
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
    Push-Location "$ServerDir"
    try {
        $env:CGO_ENABLED = "0"
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"

        if ($Target -eq "desktop") {
            Write-Host "[3/4] Compiling Pure-Go Desktop GUI binary (-H windowsgui, tray enabled)..." -ForegroundColor Yellow
            $LdFlags = "-H windowsgui -s -w"
            $Entrypoint = "./cmd/netra-uptime-desktop"
        } else {
            Write-Host "[3/4] Compiling Pure-Go Console Server binary (stdout logging)..." -ForegroundColor Yellow
            $LdFlags = "-s -w"
            $Entrypoint = "./cmd/netra-uptime"
        }

        go build -trimpath -ldflags="$LdFlags" -o "$OutputExe" $Entrypoint
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

    if ($CreateZip) {
        $OutputZip = Join-Path $RepoRoot "netra-uptime-windows.zip"
        Write-Host "Creating release archive: $OutputZip..." -ForegroundColor Yellow
        if (Test-Path "$OutputZip") {
            Remove-Item -Path "$OutputZip" -Force
        }
        Compress-Archive -Path "$OutputExe" -DestinationPath "$OutputZip" -Force
        $ZipInfo = Get-Item "$OutputZip"
        $ZipMB = [math]::Round($ZipInfo.Length / 1MB, 2)
        Write-Host " Archive: $OutputZip ($ZipMB MB)" -ForegroundColor Green
    }
} else {
    throw "Expected output binary not found at $OutputExe"
}
