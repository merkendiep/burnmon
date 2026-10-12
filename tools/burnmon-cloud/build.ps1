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

# Hooks (1.1.0 and up). The extract writes latest.jsonl with a temp file plus rename.
# SessionStart registers that file with the watcher (--watch), so every write
# fires FileChanged, and FileChanged copies it to the computer: the main path,
# no model tokens. Stop, UserPromptSubmit and SessionEnd copy too, as the 1.0
# fallback for a watcher that does not fire. Hooks of one event run in
# parallel, so the Stop copy carries the file of the turn before.
# 1.1.1 diagnostic: every extract run logs to diag.log, and FileChanged also
# runs extract.py --fired next to the copy, which logs the moment the copy
# started. UserPromptSubmit copies diag.log next to latest.jsonl (BurnMon
# reads only .jsonl there). Each copy is its own hook entry, so a missing
# diag.log can never stop the latest.jsonl copy.
$script = '${CLAUDE_PLUGIN_ROOT}/scripts/extract.py'
$watch = [ordered]@{ type = 'command'; command = 'python3'; args = @($script, '--watch'); timeout = 10 }
$fired = [ordered]@{ type = 'command'; command = 'python3'; args = @($script, '--fired'); timeout = 10 }
$extract = [ordered]@{ type = 'command'; command = 'python3'; args = @($script); timeout = 30 }
$extractThrottled = [ordered]@{ type = 'command'; command = 'python3'; args = @($script, '--throttle'); async = $true; timeout = 30 }
$copies = @(foreach ($t in $Target) {
    [ordered]@{
        type = 'mcp_tool'; server = 'remote-devices'; tool = 'device_commit_files'; timeout = 30
        input = [ordered]@{ files = @([ordered]@{ devicePath = $t; stagedPath = '/mnt/user-data/outputs/.burnmon/cloud/latest.jsonl' }) }
    }
})
$diagCopies = @(foreach ($t in $Target) {
    [ordered]@{
        type = 'mcp_tool'; server = 'remote-devices'; tool = 'device_commit_files'; timeout = 30
        input = [ordered]@{ files = @([ordered]@{ devicePath = ($t.Substring(0, $t.Length - 'latest.jsonl'.Length) + 'diag.log'); stagedPath = '/mnt/user-data/outputs/.burnmon/cloud/diag.log' }) }
    }
})
$hooks = [ordered]@{
    description = 'BurnMon cloud: writes a usage-only extract of each cloud Cowork session and copies it to BurnMon''s drop folder after each write. No prompts or content.'
    hooks = [ordered]@{
        SessionStart     = @([ordered]@{ hooks = @($watch) })
        PostToolUse      = @([ordered]@{ hooks = @($extractThrottled) })
        SubagentStop     = @([ordered]@{ hooks = @($extract) })
        Stop             = @([ordered]@{ hooks = @($extract) + $copies })
        FileChanged      = @([ordered]@{ matcher = 'latest.jsonl'; hooks = @($fired) + $copies })
        UserPromptSubmit = @([ordered]@{ hooks = $copies + $diagCopies })
        SessionEnd       = @([ordered]@{ hooks = $copies })
    }
}

$stage = Join-Path ([IO.Path]::GetTempPath()) ("burnmon-cloud-" + [guid]::NewGuid())
Copy-Item $src $stage -Recurse
# Running the tests leaves Python bytecode next to extract.py; it does not ship.
Get-ChildItem $stage -Recurse -Force -Directory -Filter '__pycache__' | Remove-Item -Recurse -Force
New-Item -ItemType Directory -Path (Join-Path $stage 'hooks') -Force | Out-Null
$json = $hooks | ConvertTo-Json -Depth 12
[IO.File]::WriteAllText((Join-Path $stage 'hooks\hooks.json'), $json, (New-Object Text.UTF8Encoding $false))

$dist = Join-Path $here 'dist'
New-Item -ItemType Directory -Path $dist -Force | Out-Null
$zip = Join-Path $dist "burnmon-cloud-$version.zip"
Remove-Item $zip -ErrorAction SilentlyContinue
# -Force lists .claude-plugin too: a '*' wildcard skips dot folders outside Windows.
Compress-Archive -Path (Get-ChildItem $stage -Force).FullName -DestinationPath $zip
Copy-Item $zip (Join-Path $dist "burnmon-cloud-$version.plugin") -Force
Remove-Item $stage -Recurse -Force
Write-Host "Built $zip and burnmon-cloud-$version.plugin for: $($Target -join ', ')"
