package system

import (
	"math/rand"
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestFindNpcSpawnPointFallsBackWhenMapRejectsAllAreaCandidates(t *testing.T) {
	ws := world.NewState()
	maps := fakeSpawnMap{passable: map[[2]int32]bool{}}

	x, y, ok := FindNpcSpawnPoint(NpcSpawnRule{
		MapID: 1,
		LocX1: 110,
		LocY1: 110,
		LocX2: 115,
		LocY2: 115,
	}, ws, maps, 0, rand.New(rand.NewSource(3)))

	if !ok {
		t.Fatal("義維初始化生成不應因為所有 tile 檢查失敗就整筆跳過")
	}
	if x < 110 || x >= 115 || y < 110 || y >= 115 {
		t.Fatalf("fallback 應留在 spawnlist 區域內，got (%d,%d)", x, y)
	}
}

func TestFindNpcSpawnPointFallsBackToOriginalPointWhenMapRejectsIt(t *testing.T) {
	ws := world.NewState()
	maps := fakeSpawnMap{passable: map[[2]int32]bool{}}

	x, y, ok := FindNpcSpawnPoint(NpcSpawnRule{
		MapID: 1,
		X:     100,
		Y:     100,
	}, ws, maps, 0, rand.New(rand.NewSource(1)))

	if !ok {
		t.Fatal("固定點生成不應因為 tile 檢查失敗就消失")
	}
	if x != 100 || y != 100 {
		t.Fatalf("固定點 fallback 應保留原始座標，got (%d,%d)", x, y)
	}
}
