package handler

import "testing"

func TestIsPolyScrollMatchesYiweiSoscPolyReelItems(t *testing.T) {
	cases := []struct {
		name   string
		itemID int32
		want   bool
	}{
		{name: "變形卷軸", itemID: 40088, want: true},
		{name: "象牙塔變形卷軸", itemID: 40096, want: true},
		{name: "受祝福的變形卷軸", itemID: 140088, want: true},
		{name: "義維 49308 是哈汀日記不是變形選單", itemID: 49308, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsPolyScroll(tc.itemID); got != tc.want {
				t.Fatalf("IsPolyScroll(%d) = %v, want %v", tc.itemID, got, tc.want)
			}
		})
	}
}
