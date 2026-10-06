[CmdletBinding()]
param(
    [string] $InstallDir
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"
[Net.ServicePointManager]::SecurityProtocol =
    [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

$repository = "Hayao0819/nth"
$apiHeaders = @{
    Accept = "application/vnd.github+json"
    "User-Agent" = "nth-installer"
    "X-GitHub-Api-Version" = "2022-11-28"
}
if ($env:GITHUB_TOKEN) {
    $apiHeaders.Authorization = "Bearer $($env:GITHUB_TOKEN)"
}

$release = Invoke-RestMethod `
    -Uri "https://api.github.com/repos/$repository/releases/latest" `
    -Headers $apiHeaders

$tag = [string] $release.tag_name
if ($tag -notmatch '^v?[0-9]') {
    throw "Could not determine the latest release tag."
}
$version = $tag -replace '^v', ''

if ([System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform(
        [System.Runtime.InteropServices.OSPlatform]::Windows
    )) {
    $os = "windows"
    $extension = ".exe"
} elseif ([System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform(
        [System.Runtime.InteropServices.OSPlatform]::Linux
    )) {
    $os = "linux"
    $extension = ""
} elseif ([System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform(
        [System.Runtime.InteropServices.OSPlatform]::OSX
    )) {
    $os = "darwin"
    $extension = ""
} else {
    throw "Unsupported operating system."
}

switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()) {
    "X64" {
        $architecture = "amd64"
    }
    "Arm64" {
        $architecture = "arm64"
    }
    default {
        throw "Unsupported architecture: $([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture)"
    }
}

$assetName = "nth_${version}_${os}_${architecture}${extension}"
$asset = @($release.assets | Where-Object name -EQ $assetName)[0]
$checksumsAsset = @($release.assets | Where-Object name -EQ "checksums.txt")[0]
if ($null -eq $asset) {
    throw "Release asset not found: $assetName"
}
if ($null -eq $checksumsAsset) {
    throw "Release asset not found: checksums.txt"
}

if ([string]::IsNullOrWhiteSpace($InstallDir)) {
    if ($env:NTH_INSTALL_DIR) {
        $InstallDir = $env:NTH_INSTALL_DIR
    } else {
        $InstallDir = Join-Path $HOME ".local/bin"
    }
}

$null = New-Item -ItemType Directory -Path $InstallDir -Force
$temporaryBinary = Join-Path $InstallDir ".nth.$([guid]::NewGuid().ToString('N')).tmp"
$temporaryChecksums = Join-Path $InstallDir ".nth-checksums.$([guid]::NewGuid().ToString('N')).tmp"

try {
    Invoke-WebRequest -UseBasicParsing -Uri $asset.browser_download_url -OutFile $temporaryBinary
    Invoke-WebRequest -UseBasicParsing -Uri $checksumsAsset.browser_download_url -OutFile $temporaryChecksums

    $checksumPattern = '^(?<hash>[0-9a-fA-F]{64})\s+\*?' + [regex]::Escape($assetName) + '$'
    $expectedChecksum = $null
    foreach ($line in [System.IO.File]::ReadLines($temporaryChecksums)) {
        $match = [regex]::Match($line, $checksumPattern)
        if ($match.Success) {
            $expectedChecksum = $match.Groups["hash"].Value.ToLowerInvariant()
            break
        }
    }
    if ($null -eq $expectedChecksum) {
        throw "Checksum for $assetName was not found."
    }

    $actualChecksum = (Get-FileHash -Algorithm SHA256 -LiteralPath $temporaryBinary).Hash.ToLowerInvariant()
    if ($actualChecksum -ne $expectedChecksum) {
        throw "Checksum verification failed."
    }

    if ($os -ne "windows") {
        & chmod 0755 $temporaryBinary
        if ($LASTEXITCODE -ne 0) {
            throw "Could not make the downloaded binary executable."
        }
    }

    $target = Join-Path $InstallDir "nth$extension"
    Move-Item -LiteralPath $temporaryBinary -Destination $target -Force
    $temporaryBinary = $null
} finally {
    if ($null -ne $temporaryBinary) {
        Remove-Item -LiteralPath $temporaryBinary -Force -ErrorAction SilentlyContinue
    }
    Remove-Item -LiteralPath $temporaryChecksums -Force -ErrorAction SilentlyContinue
}

Write-Host "Installed nth $tag to $target"
if (($env:PATH -split [System.IO.Path]::PathSeparator) -notcontains $InstallDir) {
    Write-Warning "Add $InstallDir to PATH to run nth."
}
