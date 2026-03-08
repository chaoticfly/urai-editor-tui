# install.ps1 — install urai from GitHub Releases
#
#   irm https://raw.githubusercontent.com/OWNER/urai/master/install.ps1 | iex
#
# Optional: pin a version
#   $env:VERSION = "v0.2.0"
#   irm .../install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo       = "OWNER/urai"   # ← replace OWNER with your GitHub username / org
$BinName    = "urai.exe"
$Asset      = "urai-windows-amd64.zip"
$BaseUrl    = "https://github.com/$Repo/releases"
$Version    = if ($env:VERSION) { $env:VERSION } else { "latest" }

# Install to %LOCALAPPDATA%\Programs\urai — standard per-user location,
# no admin rights required.
$InstallDir = "$env:LOCALAPPDATA\Programs\urai"

# ── Download URL ──────────────────────────────────────────────────────────────

$Url = if ($Version -eq "latest") {
    "$BaseUrl/latest/download/$Asset"
} else {
    "$BaseUrl/download/$Version/$Asset"
}

# ── Download, extract, install ────────────────────────────────────────────────

$Tmp = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $Tmp | Out-Null

try {
    Write-Host "Downloading $Asset ..."
    Invoke-WebRequest -Uri $Url -OutFile (Join-Path $Tmp $Asset) -UseBasicParsing

    Write-Host "Extracting ..."
    Expand-Archive -Path (Join-Path $Tmp $Asset) -DestinationPath $Tmp -Force

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item -Path (Join-Path $Tmp $BinName) -Destination (Join-Path $InstallDir $BinName) -Force

    Write-Host "Installed: $(Join-Path $InstallDir $BinName)"

} finally {
    Remove-Item -Recurse -Force $Tmp -ErrorAction SilentlyContinue
}

# ── Add to user PATH if not already present ───────────────────────────────────

$UserPath = [System.Environment]::GetEnvironmentVariable("PATH", "User") ?? ""
if ($UserPath -notlike "*$InstallDir*") {
    $Updated = ($UserPath.TrimEnd(";") + ";$InstallDir").TrimStart(";")
    [System.Environment]::SetEnvironmentVariable("PATH", $Updated, "User")
    # Also update the current session so the user can run it immediately.
    $env:PATH = "$env:PATH;$InstallDir"
    Write-Host "Added $InstallDir to your PATH."
}

Write-Host "Done. Run: urai"
