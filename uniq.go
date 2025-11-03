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
	if numChars <= 0 || len(s) <= numChars {
		if numChars > 0 && len(s) <= numChars {
			return ""
		}
		return s
	}
	return s[numChars:]
}

func processLines(lines []string, opts Options) []string {
	if len(lines) == 0 {
		return nil
	}

	type lineInfo struct {
		original string
		count    int
	}

	var groups []lineInfo
	var lastComparable string
	currentGroup := lineInfo{}

	for i, line := range lines {
		comparable := getComparablePart(line, opts)
		
		if i == 0 {
			lastComparable = comparable
			currentGroup = lineInfo{original: line, count: 1}
		} else if comparable == lastComparable {
			currentGroup.count++
		} else {
			groups = append(groups, currentGroup)
			lastComparable = comparable
			currentGroup = lineInfo{original: line, count: 1}
		}
	}
	groups = append(groups, currentGroup)

	var result []string
	for _, group := range groups {
		switch {
		case opts.Count:
			result = append(result, fmt.Sprintf("%d %s", group.count, group.original))
		case opts.Repeated:
			if group.count > 1 {
				result = append(result, group.original)
			}
		case opts.Unique:
			if group.count == 1 {
				result = append(result, group.original)
			}
		default:
			result = append(result, group.original)
		}
	}
	return result
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
