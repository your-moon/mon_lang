/*
 * mon_lang - native arm64 Mach-O executable writer (Apple Silicon)
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

// This writer emits a native arm64 Mach-O with mon_lang's own ad-hoc code
// signature and NO external tooling (no cc/as/ld/codesign). Modern macOS
// (Darwin 25+) refuses to exec a pure static LC_UNIXTHREAD arm64 binary and
// SIGKILLs any unsigned arm64 binary, so unlike the x86 path this one must:
//
//   - be dyld-loaded: LC_LOAD_DYLINKER + LC_MAIN + LC_LOAD_DYLIB(libSystem),
//     with empty chained fixups and an empty symtab (we import nothing; all
//     I/O is via raw `svc #0x80` syscalls);
//   - carry a valid ad-hoc CMS-less code signature that satisfies runtime AMFI
//     (which is far stricter than `codesign -v`): a CodeDirectory over 16 KiB
//     pages plus Requirements and (empty) signature slots.
//
// The exact structure was derisked byte-for-byte against `codesign`'s output;
// see the memory note "arm64 Mach-O recipe" for the gory details.
package macho

import (
	"crypto/sha256"
	"encoding/binary"
	"os"
)

type abuf struct{ b []byte }

func (w *abuf) u32(v uint32)   { w.b = binary.LittleEndian.AppendUint32(w.b, v) }
func (w *abuf) u32be(v uint32) { w.b = binary.BigEndian.AppendUint32(w.b, v) }
func (w *abuf) u64(v uint64)   { w.b = binary.LittleEndian.AppendUint64(w.b, v) }
func (w *abuf) u64be(v uint64) { w.b = binary.BigEndian.AppendUint64(w.b, v) }
func (w *abuf) nm(s string)    { n := make([]byte, 16); copy(n, s); w.b = append(w.b, n...) }
func (w *abuf) bytes(p []byte) { w.b = append(w.b, p...) }
func (w *abuf) u8(v byte)      { w.b = append(w.b, v) }
func (w *abuf) padTo(n int) {
	for len(w.b) < n {
		w.b = append(w.b, 0)
	}
}

const (
	armBase   = 0x100000000
	armVMPage = 0x4000 // arm64 segments align to 16 KiB
	armHdr    = 0x4000 // header + load commands live in page 0

	cpuARM64            = 0x0100000c
	lcSymtab            = 0x02
	lcLoadDylib         = 0x0c
	lcLoadDylinker      = 0x0e
	lcMain              = 0x80000028
	lcDyldChainedFixups = 0x80000034
	lcCodeSignature     = 0x1d
	lcBuildVersion      = 0x32
)

// ARM64Layout mirrors the x86 Layout: where ro/data land relative to code, so
// the encoder can resolve PC-relative (adrp/add) references before writing.
type ARM64Layout struct {
	CodeAddr int // vm address of the first code byte
	ROAddr   int // vm address of the ro blob (code+ro share __TEXT)
	DataAddr int // vm address of __DATA (0 if no globals)
}

// PlanARM64 computes the vm addresses for the given blob sizes.
func PlanARM64(codeLen, roLen, dataLen int) ARM64Layout {
	codeAddr := armBase + armHdr
	roAddr := codeAddr + codeLen
	textEnd := alignUp(armHdr+codeLen+roLen, armVMPage)
	dataAddr := 0
	if dataLen > 0 {
		dataAddr = armBase + textEnd
	}
	return ARM64Layout{CodeAddr: codeAddr, ROAddr: roAddr, DataAddr: dataAddr}
}

// buildArmSignature returns the ad-hoc SuperBlob covering region [0,len(region)).
func buildArmSignature(region []byte, ident string, textVMSize uint64) []byte {
	const csPage = 16384 // AMFI hashes 16 KiB pages on arm64
	codeLimit := len(region)
	nCode := (codeLimit + csPage - 1) / csPage

	req := &abuf{}
	req.u32be(0xfade0c01) // empty internal requirements set
	req.u32be(12)
	req.u32be(0)
	reqHash := sha256.Sum256(req.b)

	const nSpecial = 2 // slot -2 Requirements, slot -1 Info.plist(absent)
	id := ident + "\x00"
	const cdHdr = 88
	identOff := cdHdr
	hashesStart := identOff + len(id)
	hashOff := hashesStart + nSpecial*32

	cd := &abuf{}
	cd.u32be(0xfade0c02) // CodeDirectory
	cd.u32be(uint32(hashesStart + (nSpecial+nCode)*32))
	cd.u32be(0x00020400) // version 0x20400
	cd.u32be(0x00000002) // flags: adhoc
	cd.u32be(uint32(hashOff))
	cd.u32be(uint32(identOff))
	cd.u32be(nSpecial)
	cd.u32be(uint32(nCode))
	cd.u32be(uint32(codeLimit))
	cd.u8(32) // hashSize
	cd.u8(2)  // hashType SHA256
	cd.u8(0)  // platform
	cd.u8(14) // pageSize log2 = 16384
	cd.u32be(0)
	cd.u32be(0)          // scatterOffset
	cd.u32be(0)          // teamOffset
	cd.u32be(0)          // spare3
	cd.u64be(0)          // codeLimit64
	cd.u64be(0)          // execSegBase
	cd.u64be(textVMSize) // execSegLimit
	cd.u64be(1)          // execSegFlags: CS_EXECSEG_MAIN_BINARY
	cd.bytes([]byte(id))
	cd.bytes(reqHash[:])       // slot -2
	cd.bytes(make([]byte, 32)) // slot -1
	for i := 0; i < nCode; i++ {
		s, e := i*csPage, (i+1)*csPage
		if e > codeLimit {
			e = codeLimit
		}
		h := sha256.Sum256(region[s:e])
		cd.bytes(h[:])
	}

	const nBlobs = 3
	cdOff := 12 + nBlobs*8
	reqOff := cdOff + len(cd.b)
	wrapOff := reqOff + len(req.b)
	sb := &abuf{}
	sb.u32be(0xfade0cc0)
	sb.u32be(uint32(wrapOff + 8))
	sb.u32be(nBlobs)
	sb.u32be(0)             // CodeDirectory
	sb.u32be(uint32(cdOff))
	sb.u32be(2)             // Requirements
	sb.u32be(uint32(reqOff))
	sb.u32be(0x10000)       // CMS signature slot
	sb.u32be(uint32(wrapOff))
	sb.bytes(cd.b)
	sb.bytes(req.b)
	sb.u32be(0xfade0b01) // empty blob wrapper
	sb.u32be(8)
	return sb.b
}

// armChainedFixups builds a minimal (no-op) chained-fixups blob for nsegs
// segments — we import no symbols, so every segment's start offset is 0.
func armChainedFixups(nsegs int) []byte {
	starts := &abuf{}
	starts.u32(uint32(nsegs))
	for i := 0; i < nsegs; i++ {
		starts.u32(0)
	}
	const startsOff = 0x20
	h := &abuf{}
	h.u32(0)                                     // fixups_version
	h.u32(startsOff)                             // starts_offset
	h.u32(uint32(startsOff + len(starts.b)))     // imports_offset
	h.u32(uint32(startsOff + len(starts.b)))     // symbols_offset
	h.u32(0)                                     // imports_count
	h.u32(1)                                     // imports_format
	h.u32(0)                                     // symbols_format
	h.padTo(startsOff)
	h.bytes(starts.b)
	for len(h.b)%8 != 0 {
		h.u8(0)
	}
	return h.b
}

// WriteExecutableARM64 writes a signed native arm64 executable. entry is the
// offset of the entry point within code (normally 0).
func WriteExecutableARM64(path string, code, ro, data []byte, entry int) error {
	nsegs := 3 // __PAGEZERO, __TEXT, __LINKEDIT
	if len(data) > 0 {
		nsegs = 4
	}
	fixups := armChainedFixups(nsegs)

	textFileSize := alignUp(armHdr+len(code)+len(ro), armVMPage)
	textVMSize := uint64(textFileSize)
	entryoff := armHdr + entry

	dataOff := textFileSize
	dataFileSize := len(data)
	linkOff := dataOff
	if len(data) > 0 {
		linkOff = alignUp(dataOff+dataFileSize, armVMPage)
	}
	sigOff := alignUp(linkOff+len(fixups), 16)

	dylinker := "/usr/lib/dyld\x00"
	dylib := "/usr/lib/libSystem.B.dylib\x00"

	build := func(sig []byte) []byte {
		linkSize := (sigOff - linkOff) + len(sig)
		cmds := &abuf{}
		seg := func(name string, vmaddr, vmsize, off, fsize uint64, prot uint32) {
			cmds.u32(lcSegment64)
			cmds.u32(72)
			cmds.nm(name)
			cmds.u64(vmaddr)
			cmds.u64(vmsize)
			cmds.u64(off)
			cmds.u64(fsize)
			cmds.u32(prot)
			cmds.u32(prot)
			cmds.u32(0)
			cmds.u32(0)
		}
		seg("__PAGEZERO", 0, armBase, 0, 0, 0)
		seg("__TEXT", armBase, textVMSize, 0, uint64(textFileSize), 5) // r-x
		if len(data) > 0 {
			seg("__DATA", uint64(armBase+dataOff), uint64(alignUp(dataFileSize, armVMPage)),
				uint64(dataOff), uint64(dataFileSize), 3) // rw-
		}
		seg("__LINKEDIT", uint64(armBase+linkOff), uint64(alignUp(linkSize, armVMPage)),
			uint64(linkOff), uint64(linkSize), 1)

		cmds.u32(lcDyldChainedFixups)
		cmds.u32(16)
		cmds.u32(uint32(linkOff))
		cmds.u32(uint32(len(fixups)))

		cmds.u32(lcSymtab)
		cmds.u32(24)
		cmds.u32(uint32(linkOff))
		cmds.u32(0)
		cmds.u32(uint32(linkOff))
		cmds.u32(0)

		dlSize := alignUp(12+len(dylinker), 8)
		cmds.u32(lcLoadDylinker)
		cmds.u32(uint32(dlSize))
		cmds.u32(12)
		cmds.bytes([]byte(dylinker))
		for i := 12 + len(dylinker); i < dlSize; i++ {
			cmds.u8(0)
		}

		cmds.u32(lcBuildVersion)
		cmds.u32(24)
		cmds.u32(1)          // macOS
		cmds.u32(0x000B0000) // minos 11.0
		cmds.u32(0x000B0000) // sdk 11.0
		cmds.u32(0)

		cmds.u32(lcMain)
		cmds.u32(24)
		cmds.u64(uint64(entryoff))
		cmds.u64(0)

		dySize := alignUp(24+len(dylib), 8)
		cmds.u32(lcLoadDylib)
		cmds.u32(uint32(dySize))
		cmds.u32(24)
		cmds.u32(0)
		cmds.u32(0x00010000)
		cmds.u32(0x00010000)
		cmds.bytes([]byte(dylib))
		for i := 24 + len(dylib); i < dySize; i++ {
			cmds.u8(0)
		}

		ncmds := nsegs + 6 // segs + fixups+symtab+dylinker+buildver+main+dylib
		if sig != nil {
			cmds.u32(lcCodeSignature)
			cmds.u32(16)
			cmds.u32(uint32(sigOff))
			cmds.u32(uint32(len(sig)))
			ncmds++
		}

		out := &abuf{}
		out.u32(0xfeedfacf) // magic64
		out.u32(cpuARM64)
		out.u32(0)
		out.u32(2)                  // MH_EXECUTE
		out.u32(uint32(ncmds))
		out.u32(uint32(len(cmds.b)))
		out.u32(0x00200085) // NOUNDEFS|DYLDLINK|TWOLEVEL|PIE
		out.u32(0)
		out.bytes(cmds.b)
		if len(out.b) > armHdr {
			panic("macho arm64: load commands exceed header page")
		}
		out.padTo(armHdr)
		out.bytes(code)
		out.bytes(ro)
		out.padTo(textFileSize)
		if len(data) > 0 {
			out.bytes(data)
			out.padTo(linkOff)
		}
		out.bytes(fixups)
		out.padTo(sigOff)
		return out.b
	}

	// two-pass fixed point: the sig's datasize sits in the hashed header page,
	// so hash once with a real-sized sig, then re-hash to converge.
	pre := build(nil)
	sig := buildArmSignature(pre[:sigOff], "mon", textVMSize)
	full := build(sig)
	sig = buildArmSignature(full[:sigOff], "mon", textVMSize)
	full = append(build(sig)[:sigOff], sig...)

	return os.WriteFile(path, full, 0755)
}
