package main

import (
	"testing"
)

func TestInstruction(t *testing.T) {

	ins := "                                                                 // VPADDQ  XMM0,XMM1,XMM8"
	out := "    BYTE $0xc4; BYTE $0xc1; BYTE $0x71; BYTE $0xd4; BYTE $0xc0   // VPADDQ  XMM0,XMM1,XMM8"

	result, _ := assemble([]string{ins})

	if result[0] != out {
		t.Errorf("expected %s\ngot                     %s", out, result[0])
	}

}

func TestInstructionPresent(t *testing.T) {

	ins := "    BYTE $0xc4; BYTE $0xc1; BYTE $0x71; BYTE $0xd4; BYTE $0xc0   // VPADDQ  XMM0,XMM1,XMM8"
	out := "    BYTE $0xc4; BYTE $0xc1; BYTE $0x71; BYTE $0xd4; BYTE $0xc0   // VPADDQ  XMM0,XMM1,XMM8"

	result, _ := assemble([]string{ins})

	if result[0] != out {
		t.Errorf("expected %s\ngot                     %s", out, result[0])
	}

}

func TestInstructionWrongBytes(t *testing.T) {

	ins := "    BYTE $0xff; BYTE $0xff; BYTE $0xff; BYTE $0xff; BYTE $0xff   // VPADDQ  XMM0,XMM1,XMM8"
	out := "    BYTE $0xc4; BYTE $0xc1; BYTE $0x71; BYTE $0xd4; BYTE $0xc0   // VPADDQ  XMM0,XMM1,XMM8"

	result, _ := assemble([]string{ins})

	if result[0] != out {
		t.Errorf("expected %s\ngot                     %s", out, result[0])
	}

}

func TestInstructionWrongNumberOfBytes(t *testing.T) {

	ins := "    BYTE $0xff; BYTE $0xff;                                      // VPADDQ  XMM0,XMM1,XMM8"
	out := "    BYTE $0xc4; BYTE $0xc1; BYTE $0x71; BYTE $0xd4; BYTE $0xc0   // VPADDQ  XMM0,XMM1,XMM8"

	result, _ := assemble([]string{ins})

	if result[0] != out {
		t.Errorf("expected %s\ngot                     %s", out, result[0])
	}

}

func TestInstructionInDefine(t *testing.T) {

	ins := `    BYTE $0xc4; BYTE $0xc1; BYTE $0x71; BYTE $0xd4; BYTE $0xc0 \ // VPADDQ  XMM0,XMM1,XMM8`
	out := `    BYTE $0xc4; BYTE $0xc1; BYTE $0x71; BYTE $0xd4; BYTE $0xc0 \ // VPADDQ  XMM0,XMM1,XMM8`

	result, _ := assemble([]string{ins})

	if result[0] != out {
		t.Errorf("expected %s\ngot                     %s", out, result[0])
	}

}

func TestLongInstruction(t *testing.T) {

	inst := "                                                                 // VPALIGNR XMM8, XMM12, XMM12, 0x8"
	out1 := "    BYTE $0xc4; BYTE $0x43; BYTE $0x19; BYTE $0x0f; BYTE $0xc4   // VPALIGNR XMM8, XMM12, XMM12, 0x8"
	out2 := "                BYTE $0x08"

	result, _ := assemble([]string{inst})

	if result[0] != out1 {
		t.Errorf("expected %s\ngot                     %s", out1, result[0])
	} else if result[1] != out2 {
		t.Errorf("expected %s\ngot                     %s", out2, result[1])
	}

}

func TestLongInstructionWith2ndLine(t *testing.T) {

	ins1 := "                                                                 // VPALIGNR XMM8, XMM12, XMM12, 0x8"
	ins2 := "                BYTE $0x08; BYTE $0x19    "
	out1 := "    BYTE $0xc4; BYTE $0x43; BYTE $0x19; BYTE $0x0f; BYTE $0xc4   // VPALIGNR XMM8, XMM12, XMM12, 0x8"
	out2 := "                BYTE $0x08"

	result, _ := assemble([]string{ins1, ins2})

	if result[0] != out1 {
		t.Errorf("expected %s\ngot                     %s", out1, result[0])
	} else if result[1] != out2 {
		t.Errorf("expected %s\ngot                     %s", out2, result[1])
	}

}


func TestLongInstructionWith2ndLineInDefine(t *testing.T) {

	ins1 := `                                                               \ // VPALIGNR XMM8, XMM12, XMM12, 0x8`
	ins2 := `                BYTE $0x08                                     \`
	out1 := `    BYTE $0xc4; BYTE $0x43; BYTE $0x19; BYTE $0x0f; BYTE $0xc4 \ // VPALIGNR XMM8, XMM12, XMM12, 0x8`
	out2 := `                BYTE $0x08                                     \`

	result, _ := assemble([]string{ins1, ins2})

	if result[0] != out1 {
		t.Errorf("expected %s\ngot                     %s", out1, result[0])
	} else if result[1] != out2 {
		t.Errorf("expected %s\ngot                     %s", out2, result[1])
	}

}
