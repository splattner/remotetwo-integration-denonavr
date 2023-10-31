package denonavr

type DenonCursorControl string

const (
	DenonCursorControlUp     DenonCursorControl = "CUP"
	DenonCursorControlDown   DenonCursorControl = "CDN"
	DenonCursorControlLeft   DenonCursorControl = "CLT"
	DenonCursorControlRight  DenonCursorControl = "CRT"
	DenonCursorControlEnter  DenonCursorControl = "ENT"
	DenonCursorControlReturn DenonCursorControl = "RTN"
	DenonCursorControlMenu   DenonCursorControl = "MEN ON"
)

func (d *DenonAVR) CursorControl(cursorControl DenonCursorControl) int {
	status, _ := d.sendCommandToDevice(DenonCommandCursorControl, string(cursorControl))

	return status
}
