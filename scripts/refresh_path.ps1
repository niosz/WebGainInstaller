$m = [Environment]::GetEnvironmentVariable('Path', 'Machine')
$u = [Environment]::GetEnvironmentVariable('Path', 'User')
$all = ($m + ';' + $u) -split ';'
$seen = @{}
$r = [System.Collections.ArrayList]@()
foreach ($p in $all) {
    $t = $p.Trim()
    $k = $t.ToLower().TrimEnd('\')
    if ($t -and -not $seen[$k]) {
        $seen[$k] = 1
        $null = $r.Add($t)
    }
}
[IO.File]::WriteAllText(($env:TEMP + '\path_refresh.txt'), ($r -join ';'))
Write-Host "PATH aggiornato: $($r.Count) voci uniche"
