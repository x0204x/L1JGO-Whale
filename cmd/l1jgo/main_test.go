package main

import (
	"testing"

	"github.com/l1jgo/server/internal/data"
)

func TestCreateNpcFromTemplateCopiesSubMagicSpeedLikeJava(t *testing.T) {
	tmpl := &data.NpcTemplate{
		NpcID:         990002,
		Impl:          "L1Monster",
		HP:            10,
		INT:           18,
		AtkMagicSpeed: 1200,
		SubMagicSpeed: 1600,
	}

	npc := createNpcFromTemplate(tmpl, 100, 100, 4, 0, 30, nil)

	if npc.SubMagicSpeed != 1600 {
		t.Fatalf("Java NPC template sub_magic_speed 應帶到 runtime NPC，SubMagicSpeed=%d want=1600", npc.SubMagicSpeed)
	}
	if npc.AtkMagicSpeed != 1200 {
		t.Fatalf("Java NPC template atk_magic_speed 應帶到 runtime NPC，AtkMagicSpeed=%d want=1200", npc.AtkMagicSpeed)
	}
	if npc.Intel != 18 {
		t.Fatalf("Java NPC template intel 應帶到 runtime NPC，Intel=%d want=18", npc.Intel)
	}
}
