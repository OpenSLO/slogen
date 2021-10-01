package libs

import "github.com/fatih/color"

var (
	GoodResult  = color.New(color.FgHiGreen).Add(color.Bold).Printf
	GoodUResult = color.New(color.FgHiGreen).Add(color.Bold, color.Underline).Printf
	BadResult   = color.New(color.FgHiRed).Add(color.Bold).Printf
	BadUResult  = color.New(color.FgHiRed).Add(color.Bold, color.Underline).Printf
	BadInfo     = color.New(color.FgHiRed).Add(color.BlinkSlow).Printf
	GoodInfo    = color.New(color.FgGreen).Printf
	WarnInfo    = color.New(color.FgHiMagenta).Add(color.Italic, color.Faint).Printf
	WarnUInfo   = color.New(color.FgHiMagenta).Add(color.Underline, color.Faint).Printf
	OkInfo      = color.New(color.FgHiMagenta).Printf
)
