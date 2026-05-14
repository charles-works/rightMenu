//go:build windows

package menu

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"

	"rightmenu/internal/config"
)

const (
	hkeyCurrentUser      syscall.Handle = 0x80000001
	keyQueryValue        uint32         = 0x0001
	keySetValue          uint32         = 0x0002
	keyCreateSubKey      uint32         = 0x0004
	keyEnumerateSubKeys  uint32         = 0x0008
	regOptionNonVolatile uint32         = 0
	regSZ                uint32         = 1
	errorFileNotFound    syscall.Errno  = 2
	errorNoMoreItems     syscall.Errno  = 259
)

var (
	advapi32            = syscall.NewLazyDLL("advapi32.dll")
	procRegCreateKeyExW = advapi32.NewProc("RegCreateKeyExW")
	procRegOpenKeyExW   = advapi32.NewProc("RegOpenKeyExW")
	procRegSetValueExW  = advapi32.NewProc("RegSetValueExW")
	procRegDeleteKeyW   = advapi32.NewProc("RegDeleteKeyW")
	procRegEnumKeyExW   = advapi32.NewProc("RegEnumKeyExW")
	procRegCloseKey     = advapi32.NewProc("RegCloseKey")
)

type RegistryInstaller struct{}

func NewRegistryInstaller() RegistryInstaller { return RegistryInstaller{} }

func (RegistryInstaller) Install(cfg config.Config, pinnedExe string) error {
	plan := RenderRegistryPlan(cfg, pinnedExe)
	if err := deleteTree(hkeyCurrentUser, ParentKey+`\ExtendedSubCommandsKey\Shell`); err != nil {
		return err
	}
	parent, err := createKey(hkeyCurrentUser, ParentKey, keySetValue|keyCreateSubKey)
	if err != nil {
		return fmt.Errorf("create parent registry key: %w", err)
	}
	defer closeKey(parent)
	for name, value := range plan.Values {
		if err := setStringValue(parent, name, value); err != nil {
			return fmt.Errorf("set parent value %s: %w", name, err)
		}
	}
	extended, err := createKey(parent, `ExtendedSubCommandsKey\Shell`, keyCreateSubKey)
	if err != nil {
		return fmt.Errorf("create ExtendedSubCommandsKey shell: %w", err)
	}
	closeKey(extended)
	for _, cmd := range plan.Commands {
		keyPath := ParentKey + `\ExtendedSubCommandsKey\Shell\` + cmd.ID
		cmdKey, err := createKey(hkeyCurrentUser, keyPath, keySetValue|keyCreateSubKey)
		if err != nil {
			return fmt.Errorf("create command key %s: %w", cmd.ID, err)
		}
		if err := setStringValue(cmdKey, "MUIVerb", cmd.Title); err != nil {
			closeKey(cmdKey)
			return fmt.Errorf("set command title %s: %w", cmd.ID, err)
		}
		closeKey(cmdKey)
		commandKey, err := createKey(hkeyCurrentUser, keyPath+`\command`, keySetValue)
		if err != nil {
			return fmt.Errorf("create command value key %s: %w", cmd.ID, err)
		}
		if err := setStringValue(commandKey, "", cmd.Command); err != nil {
			closeKey(commandKey)
			return fmt.Errorf("set command value %s: %w", cmd.ID, err)
		}
		closeKey(commandKey)
	}
	return nil
}

func (RegistryInstaller) Uninstall() error {
	return deleteTree(hkeyCurrentUser, ParentKey)
}

func createKey(root syscall.Handle, path string, access uint32) (syscall.Handle, error) {
	var key syscall.Handle
	var disposition uint32
	path16, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	r0, _, _ := procRegCreateKeyExW.Call(
		uintptr(root),
		uintptr(unsafe.Pointer(path16)),
		0,
		0,
		uintptr(regOptionNonVolatile),
		uintptr(access),
		0,
		uintptr(unsafe.Pointer(&key)),
		uintptr(unsafe.Pointer(&disposition)),
	)
	if r0 != 0 {
		return 0, syscall.Errno(r0)
	}
	return key, nil
}

func openKey(root syscall.Handle, path string, access uint32) (syscall.Handle, error) {
	var key syscall.Handle
	path16, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	r0, _, _ := procRegOpenKeyExW.Call(uintptr(root), uintptr(unsafe.Pointer(path16)), 0, uintptr(access), uintptr(unsafe.Pointer(&key)))
	if r0 != 0 {
		return 0, syscall.Errno(r0)
	}
	return key, nil
}

func setStringValue(key syscall.Handle, name, value string) error {
	name16, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return err
	}
	value16, err := syscall.UTF16FromString(value)
	if err != nil {
		return err
	}
	r0, _, _ := procRegSetValueExW.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(name16)),
		0,
		uintptr(regSZ),
		uintptr(unsafe.Pointer(&value16[0])),
		uintptr(len(value16)*2),
	)
	if r0 != 0 {
		return syscall.Errno(r0)
	}
	return nil
}

func deleteTree(root syscall.Handle, path string) error {
	if err := deleteKey(root, path); err == nil || errors.Is(err, errorFileNotFound) {
		return nil
	}
	child, err := openKey(root, path, keyEnumerateSubKeys|keyQueryValue)
	if err != nil {
		if errors.Is(err, errorFileNotFound) {
			return nil
		}
		return fmt.Errorf("open registry key %s: %w", path, err)
	}
	defer closeKey(child)
	var names []string
	for index := uint32(0); ; index++ {
		buf := make([]uint16, 256)
		size := uint32(len(buf))
		err := enumKey(child, index, &buf[0], &size)
		if errors.Is(err, errorNoMoreItems) {
			break
		}
		if err != nil {
			return fmt.Errorf("enumerate subkeys %s: %w", path, err)
		}
		names = append(names, syscall.UTF16ToString(buf[:size]))
	}
	for _, name := range names {
		if err := deleteTree(root, path+`\`+name); err != nil {
			return err
		}
	}
	if err := deleteKey(root, path); err != nil && !errors.Is(err, errorFileNotFound) {
		return fmt.Errorf("delete registry key %s: %w", path, err)
	}
	return nil
}

func deleteKey(root syscall.Handle, path string) error {
	path16, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	r0, _, _ := procRegDeleteKeyW.Call(uintptr(root), uintptr(unsafe.Pointer(path16)))
	if r0 != 0 {
		return syscall.Errno(r0)
	}
	return nil
}

func enumKey(key syscall.Handle, index uint32, name *uint16, nameLen *uint32) error {
	r0, _, _ := procRegEnumKeyExW.Call(uintptr(key), uintptr(index), uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(nameLen)), 0, 0, 0, 0)
	if r0 != 0 {
		return syscall.Errno(r0)
	}
	return nil
}

func closeKey(key syscall.Handle) {
	_, _, _ = procRegCloseKey.Call(uintptr(key))
}
