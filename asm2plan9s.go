/*
 * Minio Cloud Storage, (C) 2016-2017 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// assemble assembles an array to lines into their
// resulting plan9 equivalent
func assemble(lines []string) ([]string, error) {

	var result []string

	for lineno, line := range lines {
		startsWithTab := strings.HasPrefix(line, "\t")
		line := strings.Replace(line, "\t", "    ", -1)
		fields := strings.Split(line, "//")
		if len(fields) == 2 && (startsAfterLongWordByteSequence(fields[0]) || len(fields[0]) == 65) {

			// test whether string before instruction is terminated with a backslash (so used in a #define)
			trimmed := strings.TrimSpace(fields[0])
			inDefine := len(trimmed) > 0 && string(trimmed[len(trimmed)-1]) == `\`

			sline, err := as(fields[1], lineno, len(fields[0]), inDefine)
			if err != nil {
				return result, err
			}
			if startsWithTab {
				sline = strings.Replace(sline, "    ", "\t", 1)
			}
			result = append(result, sline)
		} else {
			if startsWithTab {
				line = strings.Replace(line, "    ", "\t", 1)
			}
			result = append(result, line)
		}
	}

	return result, nil
}

// startsAfterLongWordByteSequence determines if an assembly instruction
// starts on a position after a combination of LONG, WORD, BYTE sequences
func startsAfterLongWordByteSequence(prefix string) bool {

	if len(strings.TrimSpace(prefix)) != 0 && !strings.HasPrefix(prefix, "    LONG $0x") &&
		!strings.HasPrefix(prefix, "    WORD $0x") && !strings.HasPrefix(prefix, "    BYTE $0x") {
		return false
	}

	length := 4 + len(prefix) + 1

	for objcodes := 3; objcodes <= 8; objcodes++ {

		ls, ws, bs := 0, 0, 0

		oc := objcodes

		for ; oc >= 4; oc -= 4 {
			ls++
		}
		if oc >= 2 {
			ws++
			oc -= 2
		}
		if oc == 1 {
			bs++
		}
		size := 4 + ls*(len("LONG $0x")+8) + ws*(len("WORD $0x")+4) + bs*(len("BYTE $0x")+2) + (ls+ws+bs-1)*len("; ")

		if length == size+6 { // comment starts after a space
			return true
		}
	}
	return false
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string, in io.Reader) ([]string, error) {
	if in == nil {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		in = file
	}

	var lines []string
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes the lines to the given file.
func writeLines(lines []string, path string, out io.Writer) error {
	if path != "" {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
		out = file
	}

	w := bufio.NewWriter(out)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func main() {

	file := ""
	if len(os.Args) >= 2 {
		file = os.Args[1]
	}

	var lines []string
	var err error
	if len(file) > 0 {
		fmt.Println("Processing file", file)
		lines, err = readLines(file, nil)
	} else {
		lines, err = readLines("", os.Stdin)
	}
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}

	result, err := assemble(lines)
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	err = writeLines(result, file, os.Stdout)
	if err != nil {
		log.Fatalf("writeLines: %s", err)
	}
}
