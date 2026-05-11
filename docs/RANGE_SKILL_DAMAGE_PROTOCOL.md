# S_RangeSkill 傷害欄位擴充說明

本文說明 Whale server 已採用的 `S_RangeSkill` 傷害欄位擴充格式，以及 yiwei Java server 與登入器需要配合修改的方向。

## 目的

原始 3.80C / yiwei 的 `S_RangeSkill` 只傳範圍技能動畫與命中狀態，不傳實際傷害。

原始格式每個目標只有：

```text
D targetObjectId
H hitFlag
```

Whale 擴充後，每個目標後面追加一個 `int32 damage`：

```text
D targetObjectId
H hitFlag
D damage
```

這讓登入器、客戶端 hook 或自訂 client parser 可以直接取得範圍技能對每個目標造成的實際傷害。

## 新版封包格式

`S_RangeSkill` opcode 為 `42`。

```text
offset  size  欄位
0       C     opcode = 42
1       C     actionId
2       D     casterId
6       H     casterX
8       H     casterY
10      C     heading
11      D     seq
15      H     gfxId
17      C     rangeType
18      H     reserved
20      H     targetCount
22      ...   target list
```

新版 target list 每筆為 10 bytes：

```text
offset              size  欄位
22 + i * 10 + 0    D     targetObjectId
22 + i * 10 + 4    H     hitFlag
22 + i * 10 + 6    D     damage
```

`hitFlag`：

```text
0x20 = 命中 / 有效計算
0x00 = miss / 不計算 / 被抵抗
```

`damage`：

```text
int32 little-endian
miss 或無傷害時為 0
```

封包尾端仍可能有 8-byte padding。登入器解析時必須依 `targetCount` 讀取目標資料，讀完後忽略 padding。

## 登入器解析方向

新版解析：

```pseudo
count = readH(packet, 20)

for i = 0; i < count; i++:
    base = 22 + i * 10
    targetId = readD(packet, base)
    hitFlag  = readH(packet, base + 4)
    damage   = readD(packet, base + 6)

    onRangeSkillDamage(targetId, hitFlag, damage)
```

如果登入器要同時支援原版 yiwei 與 Whale，建議做雙模式設定：

```text
rangeSkillDamageFormat = old
rangeSkillDamageFormat = whale
```

兩種格式：

```text
old   target stride = 6
whale target stride = 10
```

不建議只靠封包長度自動判斷，因為 `targetCount == 1` 時，舊格式加 padding 後可能與新版長度相同。

## 登入器轉發給原版 client 的注意事項

如果登入器只是 proxy/filter，而原始 3.80C client 沒有改 parser，不能把新版 opcode 42 原樣丟給 client。

原因是新版 damage 插在每個 target 中間：

```text
D targetObjectId
H hitFlag
D damage
```

原版 client 仍會用舊格式讀：

```text
D targetObjectId
H hitFlag
```

多目標時，原版 client 會把第一個目標的 `damage` 誤讀成下一個 `targetObjectId`。

若要讓原版 client 正常顯示動畫，登入器需要將新版封包轉回舊格式後再轉發：

```pseudo
newPacket = server packet opcode 42
count = readH(newPacket, 20)

oldPacket = copy(newPacket[0:22])

for i = 0; i < count; i++:
    base = 22 + i * 10
    targetId = readD(newPacket, base)
    hitFlag  = readH(newPacket, base + 4)
    damage   = readD(newPacket, base + 6)

    recordDamage(targetId, damage)

    oldPacket.writeD(targetId)
    oldPacket.writeH(hitFlag)

oldPacket.padTo8Bytes()
sendToClientWithUpdatedLength(oldPacket)
```

## yiwei Java server 修改方向

若朋友使用 yiwei Java server，且希望登入器也能取得範圍技能傷害，需要同步修改 Java server。

### 1. TargetStatus.java

檔案：

```text
l1j_yiwei_java/src/com/lineage/server/model/skill/TargetStatus.java
```

新增欄位與 getter/setter：

```java
private int _damage = 0;

public int getDamage() {
    return this._damage;
}

public void setDamage(final int damage) {
    this._damage = damage;
}
```

### 2. L1SkillUse.java

檔案：

```text
l1j_yiwei_java/src/com/lineage/server/model/skill/L1SkillUse.java
```

在 `runSkill()` 的 target loop 中，每輪先清除舊傷害：

```java
ts = iter.next();
cha = ts.getTarget();
ts.setDamage(0);
```

在每個目標完成 `_dmg` 計算後、`magic.commit()` 前，寫回傷害：

```java
ts.setDamage(this._dmg);

if ((this._dmg != 0) || (drainMana != 0)) {
    magic.commit(this._dmg, drainMana);
}
```

如果專案實際使用 `L1SkillUse2.java`，同樣需要套用相同修改。

### 3. S_RangeSkill.java

檔案：

```text
l1j_yiwei_java/src/com/lineage/server/serverpackets/S_RangeSkill.java
```

原本 target loop：

```java
this.writeD(targetobj);

if (target.isCalc()) {
    this.writeH(0x20);
} else {
    this.writeH(0x00);
}
```

改成：

```java
this.writeD(targetobj);

if (target.isCalc()) {
    this.writeH(0x20);
} else {
    this.writeH(0x00);
}

this.writeD(target.getDamage());
```

## 相容性提醒

- 已改 server + 已改 parser：可直接讀取每個目標 damage。
- 已改 server + 原版 client：需要登入器轉回舊格式，否則多目標範圍技能會解析錯。
- 原版 yiwei server + 新登入器：登入器必須使用 `old` 格式，否則會把 padding 或下一筆資料誤判為 damage。
- 原版 yiwei server 不會傳範圍技能實際 damage；只能取得 `targetObjectId` 與 `hitFlag`。

