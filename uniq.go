package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// Options содержит все параметры для утилиты uniq.
type Options struct {
	Count      bool
	Repeated   bool
	Unique     bool
	IgnoreCase bool
	NumFields  int
	NumChars   int
}

// --- Логика обработки строк ---

// getComparablePart возвращает часть строки для сравнения, делегируя работу вспомогательным функциям.
func getComparablePart(s string, opts Options) string {
	s = skipFields(s, opts.NumFields)
	s = skipChars(s, opts.NumChars)
	s = applyCase(s, opts.IgnoreCase)
	return s
}

// skipFields пропускает первые N полей в строке.
func skipFields(s string, numFields int) string {
	if numFields <= 0 {
		return s
	}

	fieldStartIdx := -1
	fieldCount := 0
	inField := false
	for i, r := range s {
		isSpace := r == ' ' || r == '\t'
		if !isSpace && !inField {
			inField = true
			fieldCount++
			if fieldCount > numFields {
				fieldStartIdx = i
				break
			}
		} else if isSpace {
			inField = false
		}
	}

	if fieldStartIdx != -1 {
		return s[fieldStartIdx:]
	}
	return ""
}

// skipChars пропускает первые N символов в строке.
func skipChars(s string, numChars int) string {
	if numChars <= 0 {
		return s
	}
	if len(s) > numChars {
		return s[numChars:]
	}
	return ""
}

// applyCase приводит строку к нижнему регистру, если это необходимо.
func applyCase(s string, ignoreCase bool) string {
	if ignoreCase {
		return strings.ToLower(s)
	}
	return s
}

// processLines выполняет основную логику: подсчет, фильтрацию и форматирование строк.
func processLines(lines []string, opts Options) []string {
	if len(lines) == 0 {
		return nil
	}

	lineCounts := make(map[string]int)
	originalLines := make(map[string]string)
	order := make([]string, 0)

	for _, line := range lines {
		comparablePart := getComparablePart(line, opts)
		if _, exists := lineCounts[comparablePart]; !exists {
			originalLines[comparablePart] = line
			order = append(order, comparablePart)
		}
		lineCounts[comparablePart]++
	}

	var result []string
	for _, comparablePart := range order {
		count := lineCounts[comparablePart]
		originalLine := originalLines[comparablePart]
		switch {
		case opts.Count:
			result = append(result, fmt.Sprintf("%d %s", count, originalLine))
		case opts.Repeated:
			if count > 1 {
				result = append(result, originalLine)
			}
		case opts.Unique:
			if count == 1 {
				result = append(result, originalLine)
			}
		default:
			result = append(result, originalLine)
		}
	}
	return result
}

// --- Логика инициализации и I/O ---

// parseFlags разбирает флаги командной строки и возвращает структуру Options.
func parseFlags() Options {
	opts := Options{}
	flag.BoolVar(&opts.Count, "c", false, "count occurrences")
	flag.BoolVar(&opts.Repeated, "d", false, "only print duplicate lines")
	flag.BoolVar(&opts.Unique, "u", false, "only print unique lines")
	flag.BoolVar(&opts.IgnoreCase, "i", false, "ignore case differences")
	flag.IntVar(&opts.NumFields, "f", 0, "avoid comparing the first N fields")
	flag.IntVar(&opts.NumChars, "s", 0, "avoid comparing the first N characters")
	flag.Parse()
	return opts
}

// validateOptions проверяет флаги на несовместимость и завершает программу в случае ошибки.
func validateOptions(opts Options) {
	if (opts.Count && opts.Repeated) || (opts.Count && opts.Unique) || (opts.Repeated && opts.Unique) {
		fmt.Fprintln(os.Stderr, "error: options -c, -d, -u are mutually exclusive")
		os.Exit(1)
	}
}

// setupIO настраивает ввод и вывод (файлы или stdin/stdout).
func setupIO() (io.Reader, io.WriteCloser) {
	var reader io.Reader = os.Stdin
	var writer io.WriteCloser = os.Stdout

	args := flag.Args()
	if len(args) > 0 {
		inputFile, err := os.Open(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening input file: %v\n", err)
			os.Exit(1)
		}
		reader = inputFile
	}

	if len(args) > 1 {
		outputFile, err := os.Create(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating output file: %v\n", err)
			os.Exit(1)
		}
		writer = outputFile
	}
	return reader, writer
}

// readLines читает все строки из заданного io.Reader.
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

// --- Главная функция ---

func main() {
	// 1. Конфигурация
	opts := parseFlags()
	validateOptions(opts)
	reader, writer := setupIO()
	defer writer.Close()

	// 2. Выполнение
	lines := readLines(reader)
	outputLines := processLines(lines, opts)

	// 3. Вывод
	for _, line := range outputLines {
		fmt.Fprintln(writer, line)
	}
}
