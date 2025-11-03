package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// Options содержит все параметры для утилиты uniq, полученные из флагов.
type Options struct {
	Count      bool // -c: подсчитать количество вхождений
	Repeated   bool // -d: выводить только повторяющиеся строки
	Unique     bool // -u: выводить только уникальные строки
	IgnoreCase bool // -i: игнорировать регистр
	NumFields  int  // -f: пропустить N полей в начале строки
	NumChars   int  // -s: пропустить N символов в начале строки
}

// getComparablePart возвращает часть строки для сравнения, применяя флаги -f, -s, и -i.
func getComparablePart(s string, opts Options) string {
	relevantPart := s

	// Применяем -f: пропуск полей.
	if opts.NumFields > 0 {
		fieldStartIdx := -1
		fieldCount := 0
		inField := false
		// Ищем индекс начала (N+1)-го поля.
		for i, r := range s {
			isSpace := r == ' ' || r == '\t'
			if !isSpace && !inField {
				inField = true
				fieldCount++
				if fieldCount > opts.NumFields {
					fieldStartIdx = i
					break
				}
			} else if isSpace {
				inField = false
			}
		}

		if fieldStartIdx != -1 {
			relevantPart = s[fieldStartIdx:]
		} else {
			// Если полей меньше, чем нужно пропустить, строка для сравнения пуста.
			relevantPart = ""
		}
	}

	// Применяем -s: пропуск символов в уже обработанной части строки.
	if opts.NumChars > 0 {
		if len(relevantPart) > opts.NumChars {
			relevantPart = relevantPart[opts.NumChars:]
		} else {
			relevantPart = ""
		}
	}

	// Применяем -i: игнорирование регистра.
	if opts.IgnoreCase {
		relevantPart = strings.ToLower(relevantPart)
	}

	return relevantPart
}

func main() {
	// 1. Определяем и парсим флаги командной строки.
	opts := Options{}
	flag.BoolVar(&opts.Count, "c", false, "count occurrences of lines")
	flag.BoolVar(&opts.Repeated, "d", false, "only print duplicate lines")
	flag.BoolVar(&opts.Unique, "u", false, "only print unique lines")
	flag.BoolVar(&opts.IgnoreCase, "i", false, "ignore case differences")
	flag.IntVar(&opts.NumFields, "f", 0, "avoid comparing the first N fields")
	flag.IntVar(&opts.NumChars, "s", 0, "avoid comparing the first N characters")
	flag.Parse()

	// 2. Проверяем на взаимоисключающие флаги.
	if (opts.Count && opts.Repeated) || (opts.Count && opts.Unique) || (opts.Repeated && opts.Unique) {
		fmt.Fprintln(os.Stderr, "error: options -c, -d, -u are mutually exclusive")
		fmt.Fprintln(os.Stderr, "usage: uniq [-c | -d | -u] [-i] [-f num] [-s chars] [input_file [output_file]]")
		os.Exit(1)
	}

	// 3. Настраиваем источники ввода и вывода.
	var reader io.Reader = os.Stdin
	var writer io.Writer = os.Stdout

	args := flag.Args()
	// Если передан input_file, открываем его.
	if len(args) > 0 {
		inputFile, err := os.Open(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening input file: %v\n", err)
			os.Exit(1)
		}
		defer inputFile.Close()
		reader = inputFile
	}

	// Если передан output_file, создаем его.
	if len(args) > 1 {
		outputFile, err := os.Create(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer outputFile.Close()
		writer = outputFile
	}

	// 4. Читаем все строки из источника.
	scanner := bufio.NewScanner(reader)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		os.Exit(1)
	}

	// 5. Выполняем основную логику.
	if len(lines) == 0 {
		return
	}

	lineCounts := make(map[string]int)       // Карта для подсчета вхождений.
	originalLines := make(map[string]string) // Карта для хранения оригинальной строки.
	order := make([]string, 0)               // Срез для сохранения порядка появления строк.

	for _, line := range lines {
		comparablePart := getComparablePart(line, opts)
		// Если встречаем новую (по правилам сравнения) строку, запоминаем ее.
		if _, exists := lineCounts[comparablePart]; !exists {
			originalLines[comparablePart] = line
			order = append(order, comparablePart)
		}
		lineCounts[comparablePart]++
	}

	// 6. Формируем и выводим результат.
	for _, comparablePart := range order {
		count := lineCounts[comparablePart]
		originalLine := originalLines[comparablePart]

		// Определяем, нужно ли выводить строку и в каком формате.
		switch {
		case opts.Count:
			// "-c": выводим счетчик и строку.
			fmt.Fprintf(writer, "%d %s\n", count, originalLine)
		case opts.Repeated:
			// "-d": выводим только строки, встретившиеся >1 раза.
			if count > 1 {
				fmt.Fprintln(writer, originalLine)
			}
		case opts.Unique:
			// "-u": выводим только строки, встретившиеся 1 раз.
			if count == 1 {
				fmt.Fprintln(writer, originalLine)
			}
		default:
			// По умолчанию: выводим первую встреченную уникальную строку.
			fmt.Fprintln(writer, originalLine)
		}
	}
}
