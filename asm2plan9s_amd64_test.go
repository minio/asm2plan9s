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
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestInstruction(t *testing.T) {

	ins := "                                 // VPADDQ  XMM0,XMM1,XMM8"
	out := "    LONG $0xd471c1c4; BYTE $0xc0 // VPADDQ  XMM0,XMM1,XMM8"

	result, _ := assemble([]string{ins}, false)

	if result[0] != out {
		t.Errorf("expected %s\ngot                     %s", out, result[0])
	}
}

func TestInstructionPresent(t *testing.T) {

	ins := "    LONG $0xd471c1c4; BYTE $0xc0 // VPADDQ  XMM0,XMM1,XMM8"
	out := "    LONG $0xd471c1c4; BYTE $0xc0 // VPADDQ  XMM0,XMM1,XMM8"

	result, _ := assemble([]string{ins}, false)

	if result[0] != out {
		t.Errorf("expected %s\ngot                     %s", out, result[0])
	}
}

func TestInstructionWrongBytes(t *testing.T) {

	ins := "    LONG $0x003377bb; BYTE $0xff // VPADDQ  XMM0,XMM1,XMM8"
	out := "    LONG $0xd471c1c4; BYTE $0xc0 // VPADDQ  XMM0,XMM1,XMM8"

	result, _ := assemble([]string{ins}, false)

	if result[0] != out {
		t.Errorf("expected %s\ngot                     %s", out, result[0])
	}
}

func TestInstructionInDefine(t *testing.T) {

	ins := `    LONG $0x00000000; BYTE $0xdd                               \ // VPADDQ  XMM0,XMM1,XMM8`
	out := `    LONG $0xd471c1c4; BYTE $0xc0                               \ // VPADDQ  XMM0,XMM1,XMM8`

	result, _ := assemble([]string{ins}, false)

	if result[0] != out {
		t.Errorf("expected %s\ngot                     %s", out, result[0])
	}
}

func TestCompactMultipleInstructions(t *testing.T) {

	ins1 := "                                 // VPADDQ  XMM0,XMM1,XMM8"
	ins2 := "                                 // VPADDQ  XMM1,XMM2,XMM3"
	ins3 := "                                 // VPADDQ  XMM4,XMM5,XMM6"
	ins4 := "                                 // VPADDQ  XMM4,XMM5,XMM6"
	ins5 := "                                 // VPADDQ  XMM4,XMM5,XMM6"
	ins6 := "                                 // VPADDQ  XMM4,XMM5,XMM6"
	ins7 := "                                 // VPADDQ  XMM4,XMM5,XMM6"
	ins8 := "    MOVQ AX, BX"
	ins9 := "                                 // VPADDQ  XMM0,XMM1,XMM8"
	ins10 := "                                 // VPADDQ  XMM1,XMM2,XMM3"
	ins11 := "    MOVQ BX, CX"
	ins12 := "                                 // VPADDQ  XMM4,XMM5,XMM6"
	ins13 := "                                 // VPADDQ  XMM5,XMM6,XMM0"
	out0 := "    QUAD $0xd4e9c5c0d471c1c4; QUAD $0xd4d1c5e6d4d1c5cb; QUAD $0xd4d1c5e6d4d1c5e6; QUAD $0x71c1c4e6d4d1c5e6; WORD $0xc0d4"
	out1 := "    MOVQ AX, BX"
	out2 := "    QUAD $0xe6d4d1c5cbd4e9c5"
	out3 := "    MOVQ BX, CX"
	out4 := "    LONG $0xe8d4c9c5"
	out := make([]string, 5)
	out[0], out[1], out[2], out[3], out[4] = out0, out1, out2, out3, out4

	result, _ := assemble([]string{ins1, ins2, ins3, ins4, ins5, ins6, ins7, ins8, ins9, ins10, ins11, ins12, ins13}, true)
	if len(result) != len(out) {
		t.Errorf("expected length %d\ngot             length %d", len(out), len(result))
	}
	for i := range result {
		if result[i] != out[i] {
			t.Errorf("expected %s\ngot                     %s", out[i], result[i])
		}
	}
}

func TestLongInstruction(t *testing.T) {

	ins := "                                   // VPALIGNR XMM8, XMM12, XMM12, 0x8"
	out := "    LONG $0x0f1943c4; WORD $0x08c4 // VPALIGNR XMM8, XMM12, XMM12, 0x8"

	result, _ := assemble([]string{ins}, false)

	if result[0] != out {
		t.Errorf("expected %s\ngot                     %s", out, result[0])
	}
}

func TestToPlan9sGasSingleLineEVEX(t *testing.T) {

	ins := `1                    .intel_syntax noprefix
   2 0000 62D1F548       VPADDQ  ZMM0,ZMM1,ZMM8
   2      D4C0
   3              `

	out := make([][]byte, 1, 1)
	out[0] = []byte{98, 209, 245, 72, 212, 192}

	tmpfile, err := ioutil.TempFile("", "test")
	if err != nil {
		return
	}

	if _, err := tmpfile.Write([]byte(ins)); err != nil {
		return
	}
	if err := tmpfile.Close(); err != nil {
		return
	}
	defer os.Remove(tmpfile.Name()) // clean up

	result, _ := toPlan9sGas(tmpfile.Name())
	if len(result) != len(out) || !bytes.Equal(result[0], out[0]) {
		t.Errorf("expected %v\ngot                     %v", out, result)
	}
}

func TestToPlan9sGasMultiLinesVEX(t *testing.T) {

	ins := `   1                    .intel_syntax noprefix
   2 0000 C4C171D4       VPADDQ  XMM0,XMM1,XMM8
   2      C0
   3 0005 C4C169D4       VPADDQ  XMM1,XMM2,XMM9
   3      C9
   4 000a C4C161D4       VPADDQ  XMM2,XMM3,XMM10
   4      D2
   5          `

	out := make([][]byte, 3, 3)
	out[0] = []byte{196, 193, 113, 212, 192}
	out[1] = []byte{196, 193, 105, 212, 201}
	out[2] = []byte{196, 193, 97, 212, 210}

	tmpfile, err := ioutil.TempFile("", "test")
	if err != nil {
		return
	}

	if _, err := tmpfile.Write([]byte(ins)); err != nil {
		return
	}
	if err := tmpfile.Close(); err != nil {
		return
	}
	defer os.Remove(tmpfile.Name()) // clean up

	result, _ := toPlan9sGas(tmpfile.Name())
	if len(result) != len(out) || !bytes.Equal(result[0], out[0]) ||
		!bytes.Equal(result[1], out[1]) || !bytes.Equal(result[2], out[2]) {
		t.Errorf("expected %v\ngot                     %v", out, result)
	}
}
