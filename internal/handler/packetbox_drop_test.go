package handler

import (
	"bytes"
	"testing"
)

func TestBuildItemBoardMatchesJavaFormat(t *testing.T) {
	got := BuildItemBoard(0x1234, "item")
	want := []byte{
		250, 190,
		0x34, 0x12,
		'i', 't', 'e', 'm', 0,
		0, 0, 0, 0, 0, 0, 0,
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("S_ItemBoard 封包不符 Java 格式\ngot=%v\nwant=%v", got, want)
	}
}

func TestBuildShowDropMatchesJavaFormat(t *testing.T) {
	got := BuildShowDrop(ShowDropAdena, 123456)
	want := []byte{
		250, 192, 1,
		0x40, 0xe2, 0x01, 0x00,
		0,
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("S_ShowDrop 封包不符 Java 格式\ngot=%v\nwant=%v", got, want)
	}
}
