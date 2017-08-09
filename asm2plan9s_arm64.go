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

	// as -march=armv8-a+crypto -o first.out -al=first.lis first.s
	app := "as"

	arg0 := "-march=armv8-a+crypto" // See https://gcc.gnu.org/onlinedocs/gcc-4.9.1/gcc/ARM-Options.html
	arg1 := "-o"
	arg2 := objFile
	arg3 := fmt.Sprintf("-al=%s", lisFile)
	arg4 := asmFile

	cmd := exec.Command(app, arg0, arg1, arg2, arg3, arg4)
	cmb, err := cmd.CombinedOutput()
	if err != nil {
		asmErrs := strings.Split(string(cmb)[len(asmFile)+1:], ":")
		asmErr := strings.Join(asmErrs[1:], ":")
		return "", errors.New(fmt.Sprintf("GAS error (line %d for '%s'):", lineno+1, strings.TrimSpace(instr)) + asmErr)
	}

	return toPlan9sArm(lisFile, instr)
}

func toPlan9sArm(listFile, instr string) (string, error) {

	var regexp = regexp.MustCompile(`^\s+\d+\s+\d+\s+([0-9a-fA-F]+)`)

	outputLines, err := readLines(listFile)
	if err != nil {
		return "", err
	}

	lastLine := outputLines[len(outputLines)-1]

	sline := "    "

	if match := regexp.FindStringSubmatch(lastLine); len(match) > 1 {
		sline += fmt.Sprintf("WORD $0x%s%s%s%s", strings.ToLower(match[1][6:8]), strings.ToLower(match[1][4:6]), strings.ToLower(match[1][2:4]), strings.ToLower(match[1][0:2]))
	} else {
		return "", errors.New("Regexp failed")
	}

	sline += " //" + instr

	// fmt.Println(sline)

	return sline, nil
}
