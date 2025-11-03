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
	Dups       bool
	Uniq       bool
	IgnoreCase bool
	SkipFields int
	SkipChars  int
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: uniq [-c | -d | -u] [-i] [-f num] [-s chars] [input_file [output_file]]")
}

func makeKey(line string, opt Options) string {
	s := line
	// skip fields (words separated by spaces; empty sequences ignored)
	if opt.SkipFields > 0 {
		i, fields := 0, 0
		// skip leading spaces
		for i < len(s) && s[i] == ' ' {
			i++
		}
		for i < len(s) && fields < opt.SkipFields {
			// consume word
			for i < len(s) && s[i] != ' ' {
				i++
			}
			fields++
			// consume single or multiple spaces after word
			for i < len(s) && s[i] == ' ' {
				i++
			}
		}
		if i < len(s) {
			s = s[i:]
		} else {
			s = ""
		}
	}
	// skip chars
	if opt.SkipChars > 0 {
		if opt.SkipChars < len(s) {
			s = s[opt.SkipChars:]
		} else {
			s = ""
		}
	}
	// case-insensitive
	if opt.IgnoreCase {
		s = strings.ToLower(s)
	}
	return s
}

func process(r io.Reader, w io.Writer, opt Options) error {
	sc := bufio.NewScanner(r)

	var prevLine, prevKey string
	count := 0
	flush := func() {
		if count == 0 {
			return
		}
		switch {
		case opt.Count:
			fmt.Fprintf(w, "%d %s\n", count, prevLine)
		case opt.Dups:
			if count > 1 {
				fmt.Fprintln(w, prevLine)
			}
		case opt.Uniq:
			if count == 1 {
				fmt.Fprintln(w, prevLine)
			}
		default:
			fmt.Fprintln(w, prevLine)
		}
	}

	for sc.Scan() {
		line := sc.Text()
		key := makeKey(line, opt)
		if count == 0 {
			prevLine, prevKey, count = line, key, 1
			continue
		}
		if key == prevKey {
			count++
		} else {
			flush()
			prevLine, prevKey, count = line, key, 1
		}
	}
	flush()
	if err := sc.Err(); err != nil {
		return err
	}
	return nil
}

func main() {
	var opt Options
	flag.BoolVar(&opt.Count, "c", false, "prefix lines by the number of occurrences")
	flag.BoolVar(&opt.Dups, "d", false, "only print duplicate lines")
	flag.BoolVar(&opt.Uniq, "u", false, "only print unique lines")
	flag.BoolVar(&opt.IgnoreCase, "i", false, "ignore differences in case")
	flag.IntVar(&opt.SkipFields, "f", 0, "ignore first num fields")
	flag.IntVar(&opt.SkipChars, "s", 0, "ignore first chars after fields")
	flag.Usage = usage
	flag.Parse()

	// validate mutually exclusive modes
	modes := 0
	for _, b := range []bool{opt.Count, opt.Dups, opt.Uniq} {
		if b {
			modes++
		}
	}
	if modes > 1 {
		usage()
		os.Exit(2)
	}

	args := flag.Args()
	var in io.Reader = os.Stdin
	var out io.Writer = os.Stdout
	var inFile, outFile *os.File
	var err error

	if len(args) >= 1 {
		inFile, err = os.Open(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer inFile.Close()
		in = inFile
	}
	if len(args) >= 2 {
		outFile, err = os.Create(args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer outFile.Close()
		out = outFile
	}
	if len(args) > 2 {
		usage()
		os.Exit(2)
	}

	if err := process(in, out, opt); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
