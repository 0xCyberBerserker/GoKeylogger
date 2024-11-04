# Keylogger and Clipboard Monitor

This project is a Windows-based keylogger and clipboard monitoring tool built in Go. It captures keyboard events and monitors the clipboard for changes, saving both logs to separate files. This tool demonstrates low-level interaction with Windows APIs to intercept keyboard events and read the clipboard's content.

## Features

- **Keylogging**: Captures and logs keystrokes to a file (`keystrokes.txt`).
- **Clipboard Monitoring**: Monitors changes to the clipboard content and logs each new entry to a file (`clipboard.txt`).
- **Keyboard Layout Detection**: Detects and applies specific keyboard mappings (e.g., Spanish, English, etc.).

## Requirements

- **Go**: Ensure Go is installed. You can download it [here](https://golang.org/dl/).
- **Windows OS**: This project is designed for Windows only, due to the specific use of Windows APIs.

## Setup

1. **Clone the Repository**
   ```powershell
   git clone https://github.com/0xCyberBerserker/keyloggy.git
   cd keylogger
   ```
2. **Install Dependencies**: No external Go packages are required. However, ensure Go can locate user32.dll and kernel32.dll as they are used to access Windows system APIs.

## Build and Run

   Run the following command in the project directory inside PowerShell:
   ```powershell
   .\compile.ps1
   ```


## Files Generated

- **`keystrokes.txt`**: This file logs all captured keystrokes, including special characters and key combinations.
- **`clipboard.txt`**: This file logs all unique clipboard entries detected while the program is running.

## Code Structure

- **`main.go`**: Main program file containing:
  - `monitorClipboard()`: Monitors the clipboard for changes every 2 seconds.
  - `LowLevelKeyboardProc()`: Handles keyboard events, capturing keystrokes.
  - `readClipboard()`: Reads the current content of the clipboard.

## Known Issues and Limitations

- **Permission Issues**: Some Windows versions may restrict clipboard and keyboard hook access. Run the program with Administrator privileges if you encounter errors.
- **Clipboard Check Interval**: The clipboard is checked every 2 seconds. You can modify this interval in the `monitorClipboard` function.

## License

This project is licensed under the MIT License.

---

Feel free to explore and modify the code for educational purposes. **Note**: Always ensure ethical and legal compliance when using keylogging tools.
