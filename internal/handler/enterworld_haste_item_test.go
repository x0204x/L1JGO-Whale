package handler

import (
	"testing"

	"github.com/l1jgo/server/internal/world"
)

func TestSkipRestoredMoveSpeedBuffForHasteItemLikeJava(t *testing.T) {
	player := &world.PlayerInfo{HasteItemEquipped: 1}
	for _, skillID := range []int32{29, 43, 54, 76, 152, SkillStatusHaste} {
		if !skipRestoredMoveSpeedBuffForHasteItem(player, skillID, 1) {
			t.Fatalf("yiwei haste item 進入世界時應跳過速度 buff 還原，skillID=%d", skillID)
		}
	}
	if skipRestoredMoveSpeedBuffForHasteItem(player, 21, 0) {
		t.Fatal("非速度 buff 不應因 haste item 被跳過")
	}
	if skipRestoredMoveSpeedBuffForHasteItem(&world.PlayerInfo{}, 43, 1) {
		t.Fatal("未裝備 haste item 時應允許速度 buff 還原")
	}
}
