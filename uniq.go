package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type Options struct {
	Count      bool
	Repeated   bool
	Unique     bool
	IgnoreCase bool
	NumFields  int
	NumChars   int
}

type lineGroup struct {
	original string
	count    int
}

func getComparablePart(s string, opts Options) string {
	s = skipFields(s, opts.NumFields)
	s = skipChars(s, opts.NumChars)
	if opts.IgnoreCase {
		s = strings.ToLower(s)
	}
	return s
}

func skipFields(s string, numFields int) string {
	if numFields <= 0 {
		return s
	}
	fieldCount := 0
	inField := false
	for i, r := range s {
		isSpace := r == ' ' || r == '\t'
		if !isSpace && !inField {
			inField = true
			fieldCount++
			if fieldCount > numFields {
				return s[i:]
			}
		} else if isSpace {
			inField = false
		}
	}
	return ""
}

func skipChars(s string, numChars int) string {
	if numChars <= 0 {
		return s
	}
	if len(s) > numChars {
		return s[numChars:]
	}
	return ""
}

func groupLines(lines []string, opts Options) []lineGroup {
	if len(lines) == 0 {
		return nil
	}

	var groups []lineGroup
	lastComparable := getComparablePart(lines[0], opts)
	currentGroup := lineGroup{original: lines[0], count: 1}

	for i := 1; i < len(lines); i++ {
		comparable := getComparablePart(lines[i], opts)
		if comparable == lastComparable {
			currentGroup.count++
		} else {
			groups = append(groups, currentGroup)
			lastComparable = comparable
			currentGroup = lineGroup{original: lines[i], count: 1}
		}
	}
	groups = append(groups, currentGroup)
	return groups
}

func formatOutput(groups []lineGroup, opts Options) []string {
	var result []string
	for _, group := range groups {
		if shouldOutput(group, opts) {
			result = append(result, formatGroup(group, opts))
		}
	}
	return result
}

func shouldOutput(group lineGroup, opts Options) bool {
	if opts.Repeated {
		return group.count > 1
	}
	if opts.Unique {
		return group.count == 1
	}
	return true
}

func formatGroup(group lineGroup, opts Options) string {
	if opts.Count {
		return fmt.Sprintf("%d %s", group.count, group.original)
	}
	return group.original
}

func processLines(lines []string, opts Options) []string {
	groups := groupLines(lines, opts)
	return formatOutput(groups, opts)
}

func parseFlags() Options {
	opts := Options{}
	flag.BoolVar(&opts.Count, "c", false, "count occurrences")
	flag.BoolVar(&opts.Repeated, "d", false, "only print duplicate lines")
	flag.BoolVar(&opts.Unique, "u", false, "only print unique lines")
	flag.BoolVar(&opts.IgnoreCase, "i", false, "ignore case")
	flag.IntVar(&opts.NumFields, "f", 0, "skip first N fields")
	flag.IntVar(&opts.NumChars, "s", 0, "skip first N chars")
	flag.Parse()
	return opts
}

func validateOptions(opts Options) {
	count := 0
	if opts.Count {
		count++
	}
	if opts.Repeated {
		count++
	}
	if opts.Unique {
		count++
	}
	if count > 1 {
		fmt.Fprintln(os.Stderr, "error: options -c, -d, -u are mutually exclusive")
		os.Exit(1)
	}
}

func setupIO() (io.Reader, io.WriteCloser) {
	reader := io.Reader(os.Stdin)
	writer := io.WriteCloser(os.Stdout)

	args := flag.Args()
	if len(args) > 0 {
		file, err := os.Open(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening input file: %v\n", err)
			os.Exit(1)
		}
		reader = file
	}
	if len(args) > 1 {
		file, err := os.Create(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating output file: %v\n", err)
			os.Exit(1)
		}
		writer = file
	}
	return reader, writer
}

func readLines(reader io.Reader) []string {
	var lines []string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		os.Exit(1)
	}
	return lines
}

func main() {
	opts := parseFlags()
	validateOptions(opts)
	reader, writer := setupIO()
	defer writer.Close()

	lines := readLines(reader)
	outputLines := processLines(lines, opts)

	for _, line := range outputLines {
		fmt.Fprintln(writer, line)
	}
}
