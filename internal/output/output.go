package output

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"golang.org/x/term"
)

var (
	bold    = color.New(color.Bold)
	dimmed  = color.New(color.Faint)
	green   = color.New(color.FgGreen)
	yellow  = color.New(color.FgYellow)
	red     = color.New(color.FgRed)
	cyan    = color.New(color.FgCyan)
)

// termWidth returns the current terminal width, falling back to 80.
func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

// PrintJSON marshals v to indented JSON and writes it to stdout.
func PrintJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error formatting JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// StatusColor returns a colorized status string.
func StatusColor(status string) string {
	switch status {
	case "completed":
		return green.Sprint(status)
	case "pending":
		return yellow.Sprint(status)
	case "expired", "cancelled":
		return red.Sprint(status)
	default:
		return status
	}
}

// RelativeTime formats a time.Time as a human-friendly relative string.
func RelativeTime(t time.Time) string {
	d := time.Since(t)
	if d < 0 {
		return timeUntil(-d)
	}
	return timeAgo(d)
}

// TimeUntilStr formats a future time as "in X".
func TimeUntilStr(t time.Time) string {
	d := time.Until(t)
	if d < 0 {
		return timeAgo(-d) + " ago (past)"
	}
	return "in " + timeUntil(d)
}

func timeAgo(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(math.Round(d.Minutes()))
		if m == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(math.Round(d.Hours()))
		if h == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", h)
	default:
		days := int(math.Round(d.Hours() / 24))
		if days == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	}
}

func timeUntil(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "less than a minute"
	case d < time.Hour:
		m := int(math.Round(d.Minutes()))
		if m == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		if m == 0 {
			if h == 1 {
				return "1 hour"
			}
			return fmt.Sprintf("%d hours", h)
		}
		return fmt.Sprintf("%d hours %d minutes", h, m)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}
}

// Table prints data in aligned columns. headers is the first row; rows
// contains the remaining data. Columns are separated by at least 2 spaces.
func Table(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// Compute column widths.
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	tw := termWidth()

	// Truncate last column if total width exceeds terminal.
	totalWidth := 0
	for _, w := range widths {
		totalWidth += w + 2
	}
	if totalWidth > tw && len(widths) > 0 {
		last := len(widths) - 1
		excess := totalWidth - tw
		if widths[last] > excess+3 {
			widths[last] -= excess
		}
	}

	// Print header.
	var hdr strings.Builder
	for i, h := range headers {
		hdr.WriteString(fmt.Sprintf("%-*s", widths[i]+2, h))
	}
	bold.Println(strings.TrimRight(hdr.String(), " "))

	// Print rows.
	for _, row := range rows {
		var line strings.Builder
		for i, cell := range row {
			if i >= len(widths) {
				break
			}
			if len(cell) > widths[i] {
				cell = cell[:widths[i]-1] + "~"
			}
			line.WriteString(fmt.Sprintf("%-*s", widths[i]+2, cell))
		}
		fmt.Println(strings.TrimRight(line.String(), " "))
	}
}

// FormCreated prints a success message after creating a form.
func FormCreated(id, url string) {
	green.Print("Form created: ")
	bold.Println(id)
	dimmed.Print("URL: ")
	cyan.Println(url)
}

// FormDetail prints full form details for ff status.
func FormDetail(id, title, status, recipient, url string, createdAt time.Time, expiresAt *time.Time) {
	printField("Form", id)
	printField("Title", title)
	printField("Status", StatusColor(status)+statusHint(status))
	printField("Recipient", recipient)
	printField("Created", RelativeTime(createdAt))
	if expiresAt != nil {
		printField("Expires", TimeUntilStr(*expiresAt))
	}
	printField("URL", cyan.Sprint(url))
}

func printField(label, value string) {
	bold.Printf("%-12s", label+":")
	fmt.Printf(" %s\n", value)
}

func statusHint(status string) string {
	switch status {
	case "pending":
		return dimmed.Sprint(" (waiting for response)")
	case "completed":
		return dimmed.Sprint(" (response received)")
	default:
		return ""
	}
}

// Waiting prints a waiting message.
func Waiting() {
	yellow.Println("Waiting for response...")
}

// Error prints an error message to stderr.
func Error(msg string) {
	red.Fprintf(os.Stderr, "Error: %s\n", msg)
}
