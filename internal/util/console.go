package util

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorGray   = "\033[90m"

	iconSuccess = "✓"
	iconError   = "✗"
	iconWarning = "⚠"
	iconInfo    = "ℹ"
	iconDebug   = "⚙"
	iconBuild   = "🔨"
	iconDeploy  = "🚀"
	iconRocket  = "📦"
	iconConfig  = "⚙️"
)

func InitLogger() {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}
	output.FormatLevel = func(i interface{}) string {
		var icon, color string
		level := fmt.Sprintf("%s", i)
		
		switch level {
		case "info":
			icon = iconInfo
			color = colorGreen
		case "error":
			icon = iconError
			color = colorRed
		case "warn":
			icon = iconWarning
			color = colorYellow
		case "debug":
			icon = iconDebug
			color = colorCyan
		default:
			icon = "•"
			color = colorWhite
		}
		
		return fmt.Sprintf("%s%s%s", color, icon, colorReset)
	}
	
	log.Logger = zerolog.New(output).With().Timestamp().Logger()
}

func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s%s %s\n", colorGreen, iconSuccess, colorReset, msg)
}

func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s%s %s\n", colorRed, iconError, colorReset, msg)
}

func Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s%s %s\n", colorYellow, iconWarning, colorReset, msg)
}

func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s%s %s\n", colorCyan, iconInfo, colorReset, msg)
}

func Build(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s%s %s\n", colorBlue, iconBuild, colorReset, msg)
}

func Deploy(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s%s %s\n", colorPurple, iconDeploy, colorReset, msg)
}
