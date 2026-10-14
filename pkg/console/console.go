package console

import (
	"fmt"
	"os"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var (
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle  = kernel32.NewProc("GetStdHandle")
	procWriteConsoleW = kernel32.NewProc("WriteConsoleW")
)

const stdOutputHandle = ^uintptr(10) // -11 (STD_OUTPUT_HANDLE)

// Print imprime strings com suporte nativo a UTF-16 via Win32 WriteConsoleW
func Print(s string) {
	hStdout, _, _ := procGetStdHandle.Call(uintptr(stdOutputHandle))
	if hStdout == 0 || hStdout == ^uintptr(0) {
		os.Stdout.WriteString(s)
		return
	}

	utf16Chars := utf16.Encode([]rune(s))
	if len(utf16Chars) == 0 {
		return
	}

	var written uint32
	ret, _, _ := procWriteConsoleW.Call(
		hStdout,
		uintptr(unsafe.Pointer(&utf16Chars[0])),
		uintptr(len(utf16Chars)),
		uintptr(unsafe.Pointer(&written)),
		0,
	)

	// Fallback para stdout padrão se o console estiver redirecionado (ex: pipe)
	if ret == 0 {
		os.Stdout.WriteString(s)
	}
}

// Println imprime linha com quebra de linha nativa do Windows (\r\n)
func Println(s string) {
	Print(s + "\r\n")
}

// Printf formata e imprime via WriteConsoleW
func Printf(format string, a ...interface{}) {
	Print(fmt.Sprintf(format, a...))
}

