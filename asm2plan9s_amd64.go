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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// See below for YASM support (older, no AVX512)

///////////////////////////////////////////////////////////////////////////////
//
// G A S   S U P P O R T
//
///////////////////////////////////////////////////////////////////////////////

//
// frank@hemelmeer: asm2plan9s$ more example.s
// .intel_syntax noprefix
//
//     VPANDQ   ZMM0, ZMM1, ZMM2
//
// frank@hemelmeer: asm2plan9s$ as -o example.o -al=example.lis example.s
// frank@hemelmeer: asm2plan9s$ more example.lis
// GAS LISTING example.s                   page 1
// 1                    .intel_syntax noprefix
// 2
// 3 0000 62F1F548          VPANDQ   ZMM0, ZMM1, ZMM2
// 3      DBC2
//

func as(instr string, lineno, commentPos int, inDefine bool) (string, error) {

	instrFields := strings.Split(instr, "/*")
	content := []byte(instrFields[0] + "\n")
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
	lisFile := tmpfile.Name() + ".lis"
	objFile := tmpfile.Name() + ".obj"
	os.Rename(tmpfile.Name(), asmFile)

	defer os.Remove(asmFile) // clean up
	defer os.Remove(lisFile) // clean up
	defer os.Remove(objFile) // clean up

	// as -o example.o -al=example.lis example.s
	app := "as"

	arg0 := "-o"
	arg1 := objFile
	arg2 := fmt.Sprintf("-al=%s", lisFile)
	arg3 := asmFile

	cmd := exec.Command(app, arg0, arg1, arg2, arg3)
	cmb, err := cmd.CombinedOutput()
	if err != nil {
		asmErrs := strings.Split(string(cmb)[len(asmFile)+1:], ":")
		asmErr := strings.Join(asmErrs[1:], ":")
		return "", errors.New(fmt.Sprintf("GAS error (line %d for '%s'):", lineno+1, strings.TrimSpace(instr)) + asmErr)
	}

	return toPlan9sGas(lisFile, instr, commentPos, inDefine)
}

func toPlan9sGas(objFile, instr string, commentPos int, inDefine bool) (string, error) {

	return "", nil
}

///////////////////////////////////////////////////////////////////////////////
//
// Y A S M   S U P P O R T
//
///////////////////////////////////////////////////////////////////////////////

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

// Rename to "as" for YASM support
func yasm_as(instr string, lineno, commentPos int, inDefine bool) (string, error) {

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

	return toPlan9sYasm(objFile, instr, commentPos, inDefine)
}

func toPlan9sYasm(objFile, instr string, commentPos int, inDefine bool) (string, error) {
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
			if commentPos-len(sline) > 0 {
				sline += strings.Repeat(" ", commentPos-len(sline))
			}
		} else {
			sline += " "
		}
	}

	sline += "//" + instr

	return sline, nil
}
