# install.ps1 — install urai from GitHub Releases
#
#   irm https://raw.githubusercontent.com/chaoticfly/urai-editor-tui/master/install.ps1 | iex
#
# Optional: pin a version
#   $env:VERSION = "v1.0.2"
#   irm .../install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo       = "chaoticfly/urai-editor-tui"
$BinName    = "urai.exe"
$Asset      = "urai-windows-amd64.zip"
$ApiUrl     = "https://api.github.com/repos/$Repo/releases"
$InstallDir = "$env:LOCALAPPDATA\Programs\urai"

# ── Resolve tag via GitHub API ────────────────────────────────────────────────

$Version = if ($env:VERSION) { $env:VERSION } else { $null }

if (-not $Version) {
    Write-Host "Fetching latest release ..."
    $Release = Invoke-RestMethod -Uri "$ApiUrl/latest" -UseBasicParsing
    $Version = $Release.tag_name
}

if (-not $Version) {
    Write-Error "Could not determine latest release tag."
    exit 1
}

$Url = "https://github.com/$Repo/releases/download/$Version/$Asset"

# ── Download, extract, install ────────────────────────────────────────────────

$Tmp = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $Tmp | Out-Null

try {
    Write-Host "Installing urai $Version ..."
    Invoke-WebRequest -Uri $Url -OutFile (Join-Path $Tmp $Asset) -UseBasicParsing

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
    $env:PATH = "$env:PATH;$InstallDir"
    Write-Host "Added $InstallDir to your PATH."
}

Write-Host "Done. Run: urai"
