package screen

import (
	"syscall"
)

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
)

const (
	smCXScreen = 0
	smCYScreen = 1
)

func GetPrimaryMonitorSize() (int, int) {
	w, _, _ := procGetSystemMetrics.Call(smCXScreen)
	h, _, _ := procGetSystemMetrics.Call(smCYScreen)
	return int(w), int(h)
}

func CalculateWindowSize(targetW, targetH int) (int, int) {
	monW, monH := GetPrimaryMonitorSize()
	if monW == 0 || monH == 0 {
		return targetW, targetH
	}

	maxW := int(float64(monW) * 0.90)
	maxH := int(float64(monH) * 0.90)

	if targetW <= maxW && targetH <= maxH {
		return targetW, targetH
	}

	ratio := float64(targetW) / float64(targetH)

	w := maxW
	h := int(float64(w) / ratio)

	if h > maxH {
		h = maxH
		w = int(float64(h) * ratio)
	}

	return w, h
}
