$ErrorActionPreference = "Stop"

$Repo = "naihetxi/jklz-parse-skill"
$BaseUrl = if ($env:JKLZ_INSTALL_BASE_URL) { $env:JKLZ_INSTALL_BASE_URL } else { "https://github.com/$Repo/releases/latest/download" }
$RawBaseUrl = "https://raw.githubusercontent.com/$Repo/main/cli/build"
$InstallDir = if ($env:JKLZ_INSTALL_DIR) { $env:JKLZ_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "jklz-parse" }
$ExeName = "jklz-parse.exe"

Write-Host "=========================================="
Write-Host "    jklz-parse CLI install"
Write-Host "=========================================="
Write-Host ""

$Arch = $env:PROCESSOR_ARCHITECTURE
if ($Arch -eq "AMD64") {
    $Target = "jklz-parse-windows-x64.exe"
} elseif ($Arch -eq "x86") {
    $Target = "jklz-parse-windows-x86.exe"
} else {
    throw "Unsupported Windows architecture: $Arch. Supported: x64, x86."
}

$DownloadUrl = "$BaseUrl/$Target"
$Dest = Join-Path $InstallDir $ExeName
$TmpFile = Join-Path $env:TEMP $Target

Write-Host "Detected platform: windows/$Arch"
Write-Host "Download URL: $DownloadUrl"

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TmpFile
} catch {
    if (-not $env:JKLZ_INSTALL_BASE_URL) {
        $FallbackUrl = "$RawBaseUrl/$Target"
        Write-Host "Release asset not found, trying repository binary: $FallbackUrl"
        Invoke-WebRequest -Uri $FallbackUrl -OutFile $TmpFile
    } else {
        throw
    }
}

try {
    & $TmpFile --help | Out-Null
} catch {
    Remove-Item -Force $TmpFile -ErrorAction SilentlyContinue
    throw "Downloaded binary is not executable on this platform."
}

Move-Item -Force $TmpFile $Dest

$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    $NewPath = if ([string]::IsNullOrWhiteSpace($UserPath)) { $InstallDir } else { "$UserPath;$InstallDir" }
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
    Write-Host ""
    Write-Host "Added to user PATH. Reopen PowerShell before running jklz-parse."
}

Write-Host ""
Write-Host "Install complete: $Dest"
Write-Host ""
Write-Host "Configure API before first use:"
Write-Host "   jklz-parse config --api-key YOUR_API_KEY --base-url http://192.168.42.15:15216"
Write-Host ""
Write-Host "Verify:"
Write-Host "   jklz-parse health"
