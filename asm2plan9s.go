package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
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

func yasm(instr string, lineno, commentPos int, inDefine bool) (string, error) {

	instrFields := strings.Split(instr, "/*")
	content := []byte("[bits 64]\n" + instrFields[0])
	tmpfile, err := ioutil.TempFile("", "asm2plan9s")
	if err != nil {
		return "", err
	}

	if _, err := tmpfile.Write(content); err != nil {
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
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
	cmb, err := cmd.CombinedOutput()
	if err != nil {
		yasmErrs := strings.Split(string(cmb)[len(asmFile)+1:], ":")
		yasmErr := strings.Join(yasmErrs[1:], ":")
		return "", errors.New(fmt.Sprintf("YASM error (line %d for '%s'):", lineno+1, strings.TrimSpace(instr)) + yasmErr)
	}

	return toPlan9s(objFile, instr, commentPos, inDefine)
}

func toPlan9s(objFile, instr string, commentPos int, inDefine bool) (string, error) {
	objcode, err := ioutil.ReadFile(objFile)
	if err != nil {
		return "", err
	}

	sline := "    "
	i := 0
	// First do LONGs (as many as needed)
	for ; len(objcode) >= 4; i++ {
		if i != 0 {
			sline += "; "
		}
		sline += fmt.Sprintf("LONG $0x%02x%02x%02x%02x", objcode[3], objcode[2], objcode[1], objcode[0])

		objcode = objcode[4:]
	}

	// Then do a WORD (if needed)
	if len(objcode) >= 2 {

		if i != 0 {
			sline += "; "
		}
		sline += fmt.Sprintf("WORD $0x%02x%02x", objcode[1], objcode[0])

		i++
		objcode = objcode[2:]
	}

	// And close with a BYTE (if needed)
	if len(objcode) == 1 {
		if i != 0 {
			sline += "; "
		}
		sline += fmt.Sprintf("BYTE $0x%02x", objcode[0])

		i++
		objcode = objcode[1:]
	}

	if inDefine {
		if commentPos > commentPos-2-len(sline) {
			if commentPos-2-len(sline) > 0 {
				sline += strings.Repeat(" ", commentPos-2-len(sline))
			}
		} else {
			sline += " "
		}
		sline += `\ `
	} else {
		if commentPos > len(sline) {
			if commentPos-len(sline) > 1 {
				sline += strings.Repeat(" ", commentPos-len(sline))
			}
		} else {
			sline += " "
		}
	}

	sline += "//" + instr

	return sline, nil
}

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

			sline, err := yasm(fields[1], lineno, len(fields[0]), inDefine)
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

		if length == size+2 || // comment starts after a space
			length == size+4 { // comment starts after a space, bash slash and another space
			return true
		}
	}
	return false
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
	fmt.Println("Processing", os.Args[1])
	lines, err := readLines(os.Args[1])
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}

	result, err := assemble(lines)
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	err = writeLines(result, os.Args[1])
	if err != nil {
		log.Fatalf("writeLines: %s", err)
	}
}
