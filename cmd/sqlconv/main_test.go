package main

import "testing"

func TestMergeBaseNormalNpcSpawnsKeepsOnlyNonMonsterBaseEntries(t *testing.T) {
	monsterFromYiwei := spawnEntryYAML{NpcID: 45005, MapID: 4, X: 33000, Y: 33000, Count: 10}
	baseMonster := spawnEntryYAML{NpcID: 45005, MapID: 4, X: 32000, Y: 32000, Count: 1}
	baseTeleporter := spawnEntryYAML{NpcID: 50014, MapID: 1, X: 32669, Y: 32811, Count: 1, Spread: "point"}
	basePandora := spawnEntryYAML{NpcID: 70014, MapID: 0, X: 32645, Y: 32951, Count: 1, Spread: "point"}
	baseMissingTemplate := spawnEntryYAML{NpcID: 999999, MapID: 4, X: 1, Y: 1, Count: 1}

	merged, added := mergeBaseNormalNpcSpawns(
		[]spawnEntryYAML{monsterFromYiwei},
		[]spawnEntryYAML{baseMonster, baseTeleporter, basePandora, baseMissingTemplate},
		map[int32]string{
			45005: "L1Monster",
			50014: "L1Teleporter",
			70014: "L1Merchant",
		},
	)

	if added != 2 {
		t.Fatalf("added = %d, want 2", added)
	}
	if len(merged) != 3 {
		t.Fatalf("len(merged) = %d, want 3", len(merged))
	}
	if merged[0] != monsterFromYiwei {
		t.Fatalf("yiwei monster spawn should stay first, got %+v", merged[0])
	}
	if merged[1] != baseTeleporter {
		t.Fatalf("base teleporter not preserved, got %+v", merged[1])
	}
	if merged[2] != basePandora {
		t.Fatalf("base pandora not preserved, got %+v", merged[2])
	}
}

func TestMergeBaseNormalNpcSpawnsPreservesBaseDuplicateRows(t *testing.T) {
	existing := spawnEntryYAML{NpcID: 50014, MapID: 1, X: 32669, Y: 32811, Count: 1}

	merged, added := mergeBaseNormalNpcSpawns(
		nil,
		[]spawnEntryYAML{existing, existing},
		map[int32]string{50014: "L1Teleporter"},
	)

	if added != 2 {
		t.Fatalf("added = %d, want 2", added)
	}
	if len(merged) != 2 {
		t.Fatalf("len(merged) = %d, want 2", len(merged))
	}
}
