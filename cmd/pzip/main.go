package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"projectzipper/internal/zipper"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <folder>\n", filepath.Base(os.Args[0]))
		fmt.Fprintln(flag.CommandLine.Output(), "Create a zip archive of the specified folder in the same parent directory.")
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}

	target := strings.Join(flag.Args(), " ")
	absTarget, err := filepath.Abs(target)
	if err != nil {
		exitWithError(err)
	}

	info, err := os.Stat(absTarget)
	if err != nil {
		exitWithError(err)
	}
	if !info.IsDir() {
		exitWithError(errors.New("target must be a directory"))
	}

	parent := filepath.Dir(absTarget)
	base := filepath.Base(absTarget)

	zipPath, err := zipper.NextArchiveName(parent, base)
	if err != nil {
		exitWithError(err)
	}

	printer := newProgressPrinter(absTarget)
	stats, err := zipper.ZipWithProgress(absTarget, zipPath, printer.OnProgress)
	if err != nil {
		exitWithError(err)
	}
	printer.Complete(zipPath, stats)

	fmt.Println(zipPath)
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, "pzip:", err)
	os.Exit(1)
}

type progressPrinter struct {
	source    string
	started   bool
	startTime time.Time
	total     int64
	lastLen   int
}

func newProgressPrinter(source string) *progressPrinter {
	return &progressPrinter{source: source}
}

func (p *progressPrinter) OnProgress(done, total int64) {
	if !p.started {
		p.started = true
		p.startTime = time.Now()
		p.total = total
		fmt.Fprintf(os.Stdout, "Creating archive for %s (%s)...\n", p.source, formatBytes(total))
	}

	line := p.renderLine(done, total)
	p.printLine(line)
}

func (p *progressPrinter) renderLine(done, total int64) string {
	const barWidth = 50

	filled := 0
	percent := 100.0
	if total > 0 {
		percent = (float64(done) / float64(total)) * 100
		if percent < 0 {
			percent = 0
		}
		if percent > 100 {
			percent = 100
		}
		filled = int((done * int64(barWidth)) / total)
	} else {
		// Empty directory; treat as complete.
		filled = barWidth
	}
	if filled > barWidth {
		filled = barWidth
	}

	bar := strings.Repeat("#", filled) + strings.Repeat("-", barWidth-filled)
	selapsed := time.Since(p.startTime)
	speed := "0 B/s"
	if selapsed > 0 {
		bytesPerSec := float64(done) / selapsed.Seconds()
		if bytesPerSec < 0 {
			bytesPerSec = 0
		}
		speedValue := int64(bytesPerSec + 0.5)
		speed = fmt.Sprintf("%s/s", formatBytes(speedValue))
	}

	return fmt.Sprintf("[%s] %3.0f%% (%s/%s) %s", bar, percent, formatBytes(done), formatBytes(total), speed)
}

func (p *progressPrinter) printLine(line string) {
	if pad := p.lastLen - len(line); pad > 0 {
		line += strings.Repeat(" ", pad)
	}
	fmt.Printf("\r%s", line)
	p.lastLen = len(line)
}

func (p *progressPrinter) Complete(zipPath string, stats zipper.ArchiveStats) {
	if !p.started {
		fmt.Println("No files to archive; created empty zip.")
		return
	}
	fmt.Print("\n")
	p.lastLen = 0
	zipInfo, err := os.Stat(zipPath)
	zipSize := int64(0)
	if err == nil {
		zipSize = zipInfo.Size()
	}
	fmt.Fprintf(os.Stdout, "✓ Archive complete: %s -> %s (%s source, %s archive, %d files)\n",
		p.source,
		zipPath,
		formatBytes(stats.TotalBytes),
		formatBytes(zipSize),
		stats.FileCount,
	)
}

func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}

	suffixes := []string{"KB", "MB", "GB", "TB", "PB"}
	div := float64(unit)
	exp := 0
	for n/int64(div) >= unit && exp < len(suffixes)-1 {
		div *= unit
		exp++
	}
	value := float64(n) / div
	return fmt.Sprintf("%.1f %s", value, suffixes[exp])
}
