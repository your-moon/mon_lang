/*
 * mon_lang - native link path (no external toolchain)
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package linker

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	codegen "github.com/your-moon/mon_lang/code_gen"
	"github.com/your-moon/mon_lang/encoder"
	"github.com/your-moon/mon_lang/macho"
)

// LinkNative assembles the program in-process and writes a runnable
// executable directly: encoder for machine code, macho for the container.
// No as, no cc, no libc - the built-in syscall stdlib is appended to every
// program. This is the default link path; Link() remains as the --cc
// fallback during the transition.
func (l *Linker) LinkNative(prog codegen.AsmProgram) error {
	outputDir := filepath.Dir(l.outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("гаралтын хавтас үүсгэхэд алдаа гарлаа: %v", err)
	}

	p, err := encoder.EncodeProgram(prog)
	if err != nil {
		return err
	}

	// __DATA blob: program globals then stdlib state, 8-byte slots. Slots
	// are uniform because access width comes from the instruction, not the
	// slot, and uniform slots keep every global aligned.
	data := []byte{}
	dataAddr := map[string]int{}
	addSlot := func(label string, init int64) {
		dataAddr[label] = len(data)
		data = binary.LittleEndian.AppendUint64(data, uint64(init))
	}
	for _, gv := range p.Globals {
		addSlot("d."+gv.Label, gv.InitValue)
	}
	for _, sd := range encoder.StdlibDataDefs() {
		addSlot("d."+sd.Label, sd.Init)
	}

	lay := macho.Plan(len(p.Code), len(p.ROData), len(data))
	// Convert blob offsets to deltas from the code start for rip fixups.
	finalAddr := map[string]int{}
	for label, off := range dataAddr {
		finalAddr[label] = lay.DataVM + off
	}
	if err := p.Finish(len(p.Code), finalAddr); err != nil {
		return err
	}

	return macho.WriteExecutable(l.outputFile, p.Code, p.ROData, data, 0)
}
