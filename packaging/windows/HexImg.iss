#ifndef AppVersion
#define AppVersion "0.1.0"
#endif

#ifndef AppArch
#define AppArch "x64"
#endif

#ifndef SourceDir
#define SourceDir "..\..\build\qt-release\deploy"
#endif

#ifndef OutputDir
#define OutputDir "..\..\exe"
#endif

[Setup]
AppId={{D6F18A2D-B7E4-4575-BA7B-DB1556D3154B}
AppName=HexImg
AppVersion={#AppVersion}
AppPublisher=Hexvork
AppPublisherURL=https://github.com/Hexvork/HexImg
AppSupportURL=https://github.com/Hexvork/HexImg/issues
AppUpdatesURL=https://github.com/Hexvork/HexImg/releases
DefaultDirName={autopf}\HexImg
DefaultGroupName=HexImg
DisableProgramGroupPage=yes
OutputDir={#OutputDir}
OutputBaseFilename=HexImg-windows-{#AppArch}-setup
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
SetupIconFile={#SourcePath}\HexImg.ico
#if AppArch == "arm64"
ArchitecturesAllowed=arm64
ArchitecturesInstallIn64BitMode=arm64
#else
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
#endif
UninstallDisplayIcon={app}\HexImg.exe

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Files]
Source: "{#SourceDir}\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{group}\HexImg"; Filename: "{app}\HexImg.exe"
Name: "{autodesktop}\HexImg"; Filename: "{app}\HexImg.exe"; Tasks: desktopicon

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Run]
Filename: "{app}\HexImg.exe"; Description: "{cm:LaunchProgram,HexImg}"; Flags: nowait postinstall skipifsilent
