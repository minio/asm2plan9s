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