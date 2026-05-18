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
		SubMagicSpeed: 1600,
	}

	npc := createNpcFromTemplate(tmpl, 100, 100, 4, 0, 30, nil)

	if npc.SubMagicSpeed != 1600 {
		t.Fatalf("Java NPC template sub_magic_speed 應帶到 runtime NPC，SubMagicSpeed=%d want=1600", npc.SubMagicSpeed)
	}
}
