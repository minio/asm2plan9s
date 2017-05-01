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
	"testing"
)

func TestInstruction(t *testing.T) {

	ins := "                                 // VPADDQ  XMM0,XMM1,XMM8"
	out := "    LONG $0xd471c1c4; BYTE $0xc0 // VPADDQ  XMM0,XMM1,XMM8"

	result, _ := assemble([]string{ins})

	if result[0] != out {
		t.Errorf("expected %s\ngot                     %s", out, result[0])
	}

}

func TestInstructionPresent(t *testing.T) {

	ins := "    LONG $0xd471c1c4; BYTE $0xc0 // VPADDQ  XMM0,XMM1,XMM8"
	out := "    LONG $0xd471c1c4; BYTE $0xc0 // VPADDQ  XMM0,XMM1,XMM8"

	result, _ := assemble([]string{ins})

	if result[0] != out {
		t.Errorf("expected %s\ngot                     %s", out, result[0])
	}

}

func TestInstructionWrongBytes(t *testing.T) {

	ins := "    LONG $0x003377bb; BYTE $0xff // VPADDQ  XMM0,XMM1,XMM8"
	out := "    LONG $0xd471c1c4; BYTE $0xc0 // VPADDQ  XMM0,XMM1,XMM8"

	result, _ := assemble([]string{ins})

	if result[0] != out {
		t.Errorf("expected %s\ngot                     %s", out, result[0])
	}

}

func TestInstructionInDefine(t *testing.T) {

	ins := `    LONG $0x00000000; BYTE $0xdd                               \ // VPADDQ  XMM0,XMM1,XMM8`
	out := `    LONG $0xd471c1c4; BYTE $0xc0                               \ // VPADDQ  XMM0,XMM1,XMM8`

	result, _ := assemble([]string{ins})

	if result[0] != out {
		t.Errorf("expected %s\ngot                     %s", out, result[0])
	}

}

func TestLongInstruction(t *testing.T) {

	ins := "                                   // VPALIGNR XMM8, XMM12, XMM12, 0x8"
	out := "    LONG $0x0f1943c4; WORD $0x08c4 // VPALIGNR XMM8, XMM12, XMM12, 0x8"

	result, _ := assemble([]string{ins})

	if result[0] != out {
		t.Errorf("expected %s\ngot                     %s", out, result[0])
	}
}