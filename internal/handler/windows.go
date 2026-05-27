package handler

import (
	"github.com/l1jgo/server/internal/net"
	"github.com/l1jgo/server/internal/net/packet"
	"go.uber.org/zap"
)

// HandleWindows 處理 C_WINDOWS（opcode 254）。
// Java: C_Windows.java — 多用途封包，依 type 欄位分派不同功能。
// 封包格式：[C type][依 type 而定的後續欄位]
func HandleWindows(sess *net.Session, r *packet.Reader, deps *Deps) {
	windowType := r.ReadC()

	player := deps.World.GetBySession(sess.ID)
	if player == nil {
		return
	}

	switch windowType {
	case 9:
		// Ctrl+Q 查詢限時地圖剩餘時間
		// Java: pc.sendPackets(new S_MapTimerOut(pc))
		SendMapTimerOut(sess, player)

	case 0x22:
		// 書籤排序（記憶場所順序配置）
		// Java: readBytes() → CharBookConfigReading.storeCharBookConfig
		// TODO: 實作書籤排序持久化

	case 0x27:
		// 變更書籤名稱
		// Java: readD(changeCount) → loop { readD(bookId), readS(newName) }
		// TODO: 實作書籤名稱修改

	case 6:
		// 龍門（副本傳送門）
		// Java: C_Windows.java case 6 — readD(itemObjID) + readD(selectDoor) → 消耗鑰匙 + 生成門衛
		HandleDragonDoorSelect(sess, player, r, deps)

	case 13:
		// 多用途：原版 Java 在點 NPC 時送、靜默忽略。
		// 觀察 @dynamic 對話下，<a action="cancel"> 等「關閉/取消」類按鈕被客戶端 patch
		// 處理為純本地關閉，不送 C_HACTION，但會送 C_Windows type 13（疑似 click 通知）。
		// 因此：收到 type 13 時，若該玩家有 live dialog，視為「玩家關閉了對話」並停止刷新。
		if player.LiveDialog != nil {
			player.LiveDialog = nil
			deps.Log.Info("dialog: cleared live dialog by C_Windows type 13",
				zap.Int32("char_id", player.CharID),
			)
		}

	default:
		deps.Log.Debug("C_Windows 未處理的 type",
			zap.Uint8("type", windowType),
			zap.String("角色", player.Name),
		)
	}
}
