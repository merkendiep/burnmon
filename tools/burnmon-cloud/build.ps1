# Builds the BurnMon cloud plugin for one computer.
#
# Runs in: PowerShell on the computer that runs BurnMon.
# -Target is where cloud sessions copy their usage extract to: the
# .burnmon\cloud\latest.jsonl file next to burnmon.exe. Give more than one
# target when sessions connect different folders; a cloud session can only
# write into a folder that is connected to it.
#
# Example:
#   cd C:\ZND\50_projects\burnmon\tools\burnmon-cloud
#   .\build.ps1 -Target 'C:\ZND\50_projects\burnmon\.burnmon\cloud\latest.jsonl'
#
# Output: dist\burnmon-cloud-<version>.plugin (install it under Customize >
# Plugins) and the same archive as .zip (for an organization upload).
param(
    [Parameter(Mandatory = $true)][string[]]$Target
)
$ErrorActionPreference = 'Stop'
$here = Split-Path -Parent $MyInvocation.MyCommand.Path
$src = Join-Path $here 'plugin'
$manifest = Get-Content (Join-Path $src '.claude-plugin\plugin.json') -Raw | ConvertFrom-Json
$version = $manifest.version

foreach ($t in $Target) {
    if (-not $t.EndsWith('\.burnmon\cloud\latest.jsonl')) {
        throw "Target '$t' must end in \.burnmon\cloud\latest.jsonl"
    }
}

$extract = [ordered]@{ type = 'command'; command = 'python3'; args = @('${CLAUDE_PLUGIN_ROOT}/scripts/extract.py'); timeout = 30 }
$extractAsync = [ordered]@{ type = 'command'; command = 'python3'; args = @('${CLAUDE_PLUGIN_ROOT}/scripts/extract.py'); async = $true; timeout = 30 }
$copies = @(foreach ($t in $Target) {
    [ordered]@{
        type = 'mcp_tool'; server = 'remote-devices'; tool = 'device_commit_files'; timeout = 30
        input = [ordered]@{ files = @([ordered]@{ devicePath = $t; stagedPath = '/mnt/user-data/outputs/.burnmon/cloud/latest.jsonl' }) }
    }
})
$hooks = [ordered]@{
    description = 'BurnMon cloud: writes a usage-only extract of each cloud Cowork session and copies it to BurnMon''s drop folder. No prompts or content.'
    hooks = [ordered]@{
        PostToolUse      = @([ordered]@{ hooks = @($extractAsync) })
        Stop             = @([ordered]@{ hooks = @($extract) + $copies })
        UserPromptSubmit = @([ordered]@{ hooks = $copies })
        SessionEnd       = @([ordered]@{ hooks = $copies })
    }
}

$stage = Join-Path ([IO.Path]::GetTempPath()) ("burnmon-cloud-" + [guid]::NewGuid())
Copy-Item $src $stage -Recurse
New-Item -ItemType Directory -Path (Join-Path $stage 'hooks') -Force | Out-Null
$json = $hooks | ConvertTo-Json -Depth 12
[IO.File]::WriteAllText((Join-Path $stage 'hooks\hooks.json'), $json, (New-Object Text.UTF8Encoding $false))

$dist = Join-Path $here 'dist'
New-Item -ItemType Directory -Path $dist -Force | Out-Null
$zip = Join-Path $dist "burnmon-cloud-$version.zip"
Remove-Item $zip -ErrorAction SilentlyContinue
Compress-Archive -Path (Join-Path $stage '*') -DestinationPath $zip
Copy-Item $zip (Join-Path $dist "burnmon-cloud-$version.plugin") -Force
Remove-Item $stage -Recurse -Force
Write-Host "Built $zip and burnmon-cloud-$version.plugin for: $($Target -join ', ')"
