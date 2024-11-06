package highlight

import "fmt"

func Bold(s string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", s)
}

func Blue(s string) string {
	return fmt.Sprintf("\033[38;2;0;170;255m%s\033[0m", s)
}

func Purple(s string) string {
	return fmt.Sprintf("\033[38;2;64;32;255m%s\033[0m", s)
}

func Green(s string) string {
	return fmt.Sprintf("\033[38;2;0;255;136m%s\033[0m", s)
}

func Red(s string) string {
	return fmt.Sprintf("\033[38;2;255;0;80m%s\033[0m", s)
}

func Yellow(s string) string {
	return fmt.Sprintf("\033[38;2;255;160;80m%s\033[0m", s)
}

func Gray(s string) string {
	return fmt.Sprintf("\033[38;2;8;8;8m%s\033[0m", s)
}

func ErrorLine(s string) string {
	return fmt.Sprintf("\033[4:3;58;2;255;0;80m%s\033[0m", s)
}
