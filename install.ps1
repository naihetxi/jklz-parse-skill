$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
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

$Dest = Join-Path $InstallDir $ExeName
$TmpFile = Join-Path $env:TEMP $Target
$Source = Join-Path $ScriptDir "cli\build\$Target"

Write-Host "Detected platform: windows/$Arch"
Write-Host "Binary file: $Source"

if (-not (Test-Path $Source)) {
    throw "Binary file not found: $Source. Please confirm the package contains cli\build\$Target."
}

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Copy-Item -Force $Source $TmpFile

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
    if ($env:Path -notlike "*$InstallDir*") {
        $env:Path = "$env:Path;$InstallDir"
    }
    Write-Host ""
    Write-Host "Added to user PATH."
    Write-Host "New terminals can run jklz-parse after they start."
    Write-Host "This PowerShell process has also been updated."
    Write-Host "If you launched this installer from cmd.exe or via powershell -Command,"
    Write-Host "open a new Command Prompt or run this once in the current cmd.exe window:"
    Write-Host "   set `"PATH=%PATH%;$InstallDir`""
} elseif ($env:Path -notlike "*$InstallDir*") {
    $env:Path = "$env:Path;$InstallDir"
}

Write-Host ""
Write-Host "Install complete: $Dest"
Write-Host ""
Write-Host "Configure API before first use:"
Write-Host "   jklz-parse config --api-key YOUR_API_KEY --base-url http://YOUR_HOST:PORT"
Write-Host ""
Write-Host "Verify:"
Write-Host "   jklz-parse health"
