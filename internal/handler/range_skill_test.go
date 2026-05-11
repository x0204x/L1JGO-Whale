package handler

import (
	"encoding/binary"
	"testing"

	"github.com/l1jgo/server/internal/net/packet"
)

func TestBuildRangeSkillWritesDamageAfterEachTargetHitFlag(t *testing.T) {
	data := BuildRangeSkill(
		1001,
		32767,
		32768,
		3,
		1804,
		18,
		RangeSkillTypeDir,
		[]RangeSkillTarget{
			{ObjectID: 2001, Hit: true, Damage: 321},
			{ObjectID: 2002, Hit: false, Damage: 0},
		},
	)

	if data[0] != packet.S_OPCODE_RANGESKILLS {
		t.Fatalf("範圍技能封包 opcode 錯誤：got=%d want=%d", data[0], packet.S_OPCODE_RANGESKILLS)
	}

	const headerLen = 22
	const targetLen = 10
	const paddingToEightBytes = 6
	wantLen := headerLen + targetLen*2 + paddingToEightBytes
	if len(data) != wantLen {
		t.Fatalf("範圍技能封包長度錯誤：got=%d want=%d data=%v", len(data), wantLen, data)
	}
	if got := binary.LittleEndian.Uint16(data[20:22]); got != 2 {
		t.Fatalf("範圍技能目標數錯誤：got=%d want=2", got)
	}

	first := data[22:32]
	if got := int32(binary.LittleEndian.Uint32(first[0:4])); got != 2001 {
		t.Fatalf("第一個目標 objectID 錯誤：got=%d want=2001", got)
	}
	if got := binary.LittleEndian.Uint16(first[4:6]); got != 0x20 {
		t.Fatalf("第一個目標 hit flag 錯誤：got=0x%04x want=0x0020", got)
	}
	if got := int32(binary.LittleEndian.Uint32(first[6:10])); got != 321 {
		t.Fatalf("第一個目標 damage 錯誤：got=%d want=321", got)
	}

	second := data[32:42]
	if got := int32(binary.LittleEndian.Uint32(second[0:4])); got != 2002 {
		t.Fatalf("第二個目標 objectID 錯誤：got=%d want=2002", got)
	}
	if got := binary.LittleEndian.Uint16(second[4:6]); got != 0 {
		t.Fatalf("第二個目標 hit flag 錯誤：got=0x%04x want=0", got)
	}
	if got := int32(binary.LittleEndian.Uint32(second[6:10])); got != 0 {
		t.Fatalf("第二個目標 damage 錯誤：got=%d want=0", got)
	}
}
