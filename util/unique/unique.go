/*
 * mon_lang - util/unique
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package unique

import "fmt"

type UniqueGen struct {
	TempCount  uint64
	LabelCount uint64
}

func NewUniqueGen() UniqueGen {
	return UniqueGen{
		TempCount:  0,
		LabelCount: 0,
	}
}

func (c *UniqueGen) MakeTemp() string {
	temp := fmt.Sprintf("tmp.%d", c.TempCount)
	c.TempCount += 1
	return temp

}

func (c *UniqueGen) MakeLabel(prefix string) string {
	temp := fmt.Sprintf("%s.%d", prefix, c.LabelCount)
	c.LabelCount += 1
	return temp

}
