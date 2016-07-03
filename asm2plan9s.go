package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"regexp"
)

//
// yasm-assemble-disassemble-roundtrip-sse.txt
//
// franks-mbp:sse frankw$ more assembly.asm
// [bits 64]
//
// VPXOR   YMM4, YMM2, YMM3    ; X4: Result
// franks-mbp:sse frankw$ yasm assembly.asm
// franks-mbp:sse frankw$ hexdump -C assembly
// 00000000  c5 ed ef e3                                       |....|
// 00000004
// franks-mbp:sse frankw$ echo 'lbl: db 0xc5, 0xed, 0xef, 0xe3' | yasm -f elf - -o assembly.o
// franks-mbp:sse frankw$ gobjdump -d -M intel assembly.o
//
// assembly.o:     file format elf32-i386
//
//
// Disassembly of section .text:
//
// 00000000 <.text>:
// 0:   c5 ed ef e3             vpxor  ymm4,ymm2,ymm3

func yasm(instr string, inDefine bool) ([]string, error) {

	instrFields := strings.Split(instr, "/*")
	content := []byte("[bits 64]\n" + instrFields[0])
	tmpfile, err := ioutil.TempFile("", "asm2plan9s")
	if err != nil {
		return []string{""}, err
	}

	if _, err := tmpfile.Write(content); err != nil {
		return []string{""}, err
	}
	if err := tmpfile.Close(); err != nil {
		return []string{""}, err
	}

	asmFile := tmpfile.Name() + ".asm"
	objFile := tmpfile.Name() + ".obj"
	os.Rename(tmpfile.Name(), asmFile)

	defer os.Remove(asmFile) // clean up
	defer os.Remove(objFile) // clean up

	app := "yasm"

	arg0 := "-o"
	arg1 := objFile
	arg2 := asmFile

	cmd := exec.Command(app, arg0, arg1, arg2)
	_, err = cmd.Output()
	if err != nil {
		return []string{""}, err
	}

	return toPlan9s(objFile, instr, inDefine)
}

func toPlan9s(objFile, instr string, inDefine bool) ([]string, error) {
	objcode, err := ioutil.ReadFile(objFile)
	if err != nil {
		return []string{""}, err
	}

	sline := "    "
	var i int
	var b byte
	for i, b = range objcode {
		if i != 0 {
			sline += "; "
		}

		sline += fmt.Sprintf("BYTE $0x%02x", b)

		if i == 4 {
			break
		}
	}

	if inDefine {
		sline += strings.Repeat(" ", 63-len(sline))
		sline += `\ `
	} else {
		sline += strings.Repeat(" ", 65-len(sline))
	}

	sline += "//" + instr

	if i < len(objcode)-1 {
		slineCtnd := "    "
		slineCtnd += "            " // additional indent for first code
		j := i + 1                  // current byte already output
		for ; j < len(objcode); j++ {
			if j != i+1 {
				slineCtnd += "; "
			}

			slineCtnd += fmt.Sprintf("BYTE $0x%02x", objcode[j])
		}

		if inDefine {
			slineCtnd += strings.Repeat(" ", 63-len(slineCtnd))
			slineCtnd += `\`
		}

		return []string{sline, slineCtnd}, nil
	}

	return []string{sline}, nil
}

// assemble assembles an array to lines into their
// resulting plan9 equivalent
func assemble(lines []string) ([]string, error) {

	var result []string

	lines, err := filterContinuedByteSequences(lines)
	if err != nil {
		return result, err
	}

	for _, line := range lines {
		line := strings.Replace(line, "\t", "    ", -1)
		fields := strings.Split(line, "//")
		if len(fields[0]) == 65 && len(fields) == 2 {

			// test whether string before instruction is terminated with a backslash (so used in a #define)
			trimmed := strings.TrimSpace(fields[0])
			inDefine := len(trimmed) > 0 && string(trimmed[len(trimmed)-1]) == `\`

			sline, err := yasm(fields[1], inDefine)
			if err != nil {
				return result, err
			}
			result = append(result, sline...)
		} else {
			result = append(result, line)
		}
	}

	return result, nil
}

// filterContinuedByteSequences filters out (on next line) continued BYTE
// sequences (for instructions that result in longer than 5 opcodes
func filterContinuedByteSequences(lines []string) ([]string, error) {

	reTwoBytes := regexp.MustCompile("[0-9a-fA-F][0-9a-fA-F]")
	reEndsWithBackslash := regexp.MustCompile(`\s*\\`)

	var filtered []string

	for _, line := range lines {

		// check prefix
		prefix := "                BYTE $0x"
		lineHexBytes := strings.HasPrefix(line, prefix)

		if lineHexBytes {
			l := strings.TrimSpace(line[len(prefix):])

			for ; len(l) > 0; {

				// check two hex characters
				lineHexBytes = reTwoBytes.FindString(l) != ""
				if !lineHexBytes {
					break
				}

				// skip and check if done
				l = l[2:len(l)]
				if len(l) == 0 {
					break
				}

				// test for string between hex codes
				interfix := "; BYTE $0x"
				lineHexBytes = strings.HasPrefix(l, interfix)
				if !lineHexBytes {
					// No more bytes, do one more check before breaking out if the statement ends with a delimiter
					lineHexBytes = reEndsWithBackslash.FindString(l) != ""
					break
				}

				// skip string inbetween to next hex
				l = l[len(interfix):]
			}
		}

		if !lineHexBytes {
			filtered = append(filtered, line)
		}
	}
	return filtered, nil
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes the lines to the given file.
func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func main() {

	if len(os.Args) < 2 {
		fmt.Printf("error: no input specified\n\n")
		fmt.Println("usage: asm2plan9s file")
		fmt.Println("  will in-place update the assembly file with proper BYTE sequence as generated by YASM")
		return
	}
	fmt.Println(os.Args[1])
	lines, err := readLines(os.Args[1])
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}

	result, err := assemble(lines)
	if err != nil {
		log.Fatalf("assemble: %s", err)
	}

	err = writeLines(result, os.Args[1])
	if err != nil {
		log.Fatalf("writeLines: %s", err)
	}
}
