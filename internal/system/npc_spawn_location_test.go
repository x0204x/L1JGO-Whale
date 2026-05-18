package system

import (
	"math/rand"
	"testing"

	"github.com/l1jgo/server/internal/world"
)

type fakeSpawnMap struct {
	passable map[[2]int32]bool
}

func (m fakeSpawnMap) IsInMap(_ int16, x, y int32) bool {
	return x >= 90 && x <= 130 && y >= 90 && y <= 130
}

func (m fakeSpawnMap) IsPassablePoint(_ int16, x, y int32) bool {
	return m.passable[[2]int32{x, y}]
}

func TestFindNpcSpawnPointUsesAreaBounds(t *testing.T) {
	ws := world.NewState()
	maps := fakeSpawnMap{passable: map[[2]int32]bool{
		{112, 111}: true,
	}}
	rng := rand.New(rand.NewSource(3))

	x, y, ok := FindNpcSpawnPoint(NpcSpawnRule{
		MapID: 1,
		X:     100,
		Y:     100,
		LocX1: 110,
		LocY1: 110,
		LocX2: 115,
		LocY2: 115,
	}, ws, maps, 0, rng)

	if !ok {
		t.Fatal("應該找到矩形範圍內可生成座標")
	}
	if x < 110 || x >= 115 || y < 110 || y >= 115 {
		t.Fatalf("生成座標應落在 Java locx1~locx2 半開區間，got (%d,%d)", x, y)
	}
	if !maps.IsPassablePoint(1, x, y) {
		t.Fatalf("生成座標必須可通行，got (%d,%d)", x, y)
	}
}

func TestFindNpcSpawnPointRejectsBlockedCenterAndFallsBack(t *testing.T) {
	ws := world.NewState()
	maps := fakeSpawnMap{passable: map[[2]int32]bool{
		{101, 100}: true,
	}}

	x, y, ok := FindNpcSpawnPoint(NpcSpawnRule{
		MapID: 1,
		X:     100,
		Y:     100,
	}, ws, maps, 0, rand.New(rand.NewSource(1)))

	if !ok {
		t.Fatal("中心點不可通行時應尋找附近可通行座標")
	}
	if x != 101 || y != 100 {
		t.Fatalf("應選擇附近第一個可通行座標，got (%d,%d)", x, y)
	}
}

func TestFindNpcSpawnPointRejectsOccupiedTile(t *testing.T) {
	ws := world.NewState()
	ws.AddNpc(&world.NpcInfo{ID: 200_000_001, X: 100, Y: 100, MapID: 1})
	maps := fakeSpawnMap{passable: map[[2]int32]bool{
		{100, 100}: true,
		{99, 100}:  true,
	}}

	x, y, ok := FindNpcSpawnPoint(NpcSpawnRule{
		MapID: 1,
		X:     100,
		Y:     100,
	}, ws, maps, 0, rand.New(rand.NewSource(1)))

	if !ok {
		t.Fatal("中心點被佔用時應尋找附近空座標")
	}
	if x == 100 && y == 100 {
		t.Fatal("不可選擇已被其他實體佔用的座標")
	}
	if !maps.IsPassablePoint(1, x, y) {
		t.Fatalf("替代座標必須可通行，got (%d,%d)", x, y)
	}
}
