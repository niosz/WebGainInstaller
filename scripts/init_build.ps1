$projectDir = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$buildWin = Join-Path $projectDir 'build\windows'

if (-not (Test-Path $buildWin)) {
    New-Item -ItemType Directory -Path $buildWin -Force | Out-Null
}

$manifest = @'
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly manifestVersion="1.0" xmlns="urn:schemas-microsoft-com:asm.v1" xmlns:asmv3="urn:schemas-microsoft-com:asm.v3">
    <assemblyIdentity type="win32" name="com.wails.{{.Name}}" version="{{.Info.ProductVersion}}.0" processorArchitecture="*"/>
    <dependency>
        <dependentAssembly>
            <assemblyIdentity type="win32" name="Microsoft.Windows.Common-Controls" version="6.0.0.0" processorArchitecture="*" publicKeyToken="6595b64144ccf1df" language="*"/>
        </dependentAssembly>
    </dependency>
    <trustInfo xmlns="urn:schemas-microsoft-com:asm.v3">
        <security>
            <requestedPrivileges>
                <requestedExecutionLevel level="requireAdministrator" uiAccess="false"/>
            </requestedPrivileges>
        </security>
    </trustInfo>
    <asmv3:application>
        <asmv3:windowsSettings>
            <dpiAware xmlns="http://schemas.microsoft.com/SMI/2005/WindowsSettings">true/pm</dpiAware>
            <dpiAwareness xmlns="http://schemas.microsoft.com/SMI/2016/WindowsSettings">permonitorv2,permonitor</dpiAwareness>
        </asmv3:windowsSettings>
    </asmv3:application>
</assembly>
'@

$infoJson = @'
{
	"fixed": {
		"file_version": "{{.Info.ProductVersion}}"
	},
	"info": {
		"0000": {
			"ProductVersion": "{{.Info.ProductVersion}}",
			"CompanyName": "{{.Info.CompanyName}}",
			"FileDescription": "{{.Info.ProductName}}",
			"LegalCopyright": "{{.Info.Copyright}}",
			"ProductName": "{{.Info.ProductName}}",
			"Comments": "{{.Info.Comments}}"
		}
	}
}
'@

[IO.File]::WriteAllText((Join-Path $buildWin 'wails.exe.manifest'), $manifest)
[IO.File]::WriteAllText((Join-Path $buildWin 'info.json'), $infoJson)

Copy-Item (Join-Path $projectDir 'media\icon.ico') (Join-Path $buildWin 'icon.ico') -Force

Copy-Item (Join-Path $projectDir 'media\icon.png') (Join-Path $projectDir 'build\appicon.png') -Force

Write-Host "build\windows\ inizializzata (manifest UAC, info.json, icon.ico, appicon.png da ico)"
