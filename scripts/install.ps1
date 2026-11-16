# Installs the Square Cloud CLI for the current user, without administrator
# rights, into %LocalAppData%\squarecloud\bin (or $env:SQUARECLOUD_INSTALL_DIR)
# and adds that folder to the user PATH:
#
#   irm https://raw.githubusercontent.com/squarecloudofc/cli/master/scripts/install.ps1 | iex

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue' # Invoke-WebRequest is far slower with its progress bar
[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

$dir = if ($env:SQUARECLOUD_INSTALL_DIR) { $env:SQUARECLOUD_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'squarecloud\bin' }

# A 32-bit PowerShell on 64-bit Windows reports x86 here and the real one in PROCESSOR_ARCHITEW6432.
$cpu = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
$arch = switch ($cpu) {
    'AMD64' { 'amd64' }
    'ARM64' { 'arm64' }
    'x86' { '386' }
    default { throw "Unsupported processor architecture: $cpu" }
}

$asset = "squarecloud_windows_$arch.zip"
$base = 'https://github.com/squarecloudofc/cli/releases/latest/download'
$tmp = Join-Path ([IO.Path]::GetTempPath()) ([Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $tmp | Out-Null

try {
    Write-Host "> Downloading $asset..."
    Invoke-WebRequest "$base/$asset" -OutFile "$tmp\$asset" -UseBasicParsing
    Invoke-WebRequest "$base/checksums.txt" -OutFile "$tmp\checksums.txt" -UseBasicParsing

    $entry = Get-Content "$tmp\checksums.txt" | Where-Object { ($_ -split '\s+')[1] -eq $asset } | Select-Object -First 1
    if (-not $entry) { throw "checksums.txt has no entry for $asset" }
    $expected = ($entry -split '\s+')[0]
    $actual = (Get-FileHash "$tmp\$asset" -Algorithm SHA256).Hash
    if ($actual -ne $expected) { throw "Checksum verification failed for $asset" } # -ne ignores case

    Expand-Archive "$tmp\$asset" -DestinationPath "$tmp\unpacked" -Force
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
    Copy-Item "$tmp\unpacked\squarecloud.exe" (Join-Path $dir 'squarecloud.exe') -Force
}
finally {
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
}

$version = & (Join-Path $dir 'squarecloud.exe') --version
Write-Host "$([char]0x2713) Installed $version in $dir"

# Read and write the raw value: going through [Environment] would expand
# entries such as %USERPROFILE%\bin for good.
$envKey = Get-Item 'HKCU:\Environment'
$userPath = $envKey.GetValue('Path', '', 'DoNotExpandEnvironmentNames')
if (($userPath -split ';') -notcontains $dir) {
    $entries = @($userPath -split ';' | Where-Object { $_ }) + $dir
    Set-ItemProperty 'HKCU:\Environment' -Name Path -Value ($entries -join ';') -Type ExpandString
    # Setting any user variable through .NET tells Explorer the environment changed.
    [Environment]::SetEnvironmentVariable('SQUARECLOUD_INSTALL', '1', 'User')
    [Environment]::SetEnvironmentVariable('SQUARECLOUD_INSTALL', $null, 'User')
    $env:Path = "$env:Path;$dir"
    Write-Host "Added $dir to your user PATH. Open a new terminal to use squarecloud."
}

Write-Host 'Optional: squarecloud integration install adds "Send to Square Cloud" to the Explorer menu.'
