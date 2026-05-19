package roomutil

import (
	"testing"

	"owl-common/card"
)

func TestClassifyRoomType(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"Bathroom", card.RoomTypeBathroom},
		{"BATHROOM 1F", card.RoomTypeBathroom},
		{"WC", card.RoomTypeBathroom},
		{"wc1F", card.RoomTypeBathroom},
		{"Restroom A", card.RoomTypeBathroom},
		{"Toilet", card.RoomTypeBathroom},
		// v2 简化：Bedroom / LivingRoom / Garage / 空名 全合并 Default
		{"Master Bedroom", card.RoomTypeDefault},
		{"bed room 2", card.RoomTypeDefault},
		{"Kitchen", card.RoomTypeKitchen},
		{"Living Room", card.RoomTypeDefault},
		{"", card.RoomTypeDefault},
		{"Garage", card.RoomTypeDefault},
	}
	for _, c := range cases {
		if got := ClassifyRoomType(c.in); got != c.want {
			t.Errorf("ClassifyRoomType(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestIsBathroom(t *testing.T) {
	if !IsBathroom("BATHROOM") {
		t.Error("IsBathroom(BATHROOM) should be true")
	}
	if IsBathroom("Living Room") {
		t.Error("IsBathroom(Living Room) should be false")
	}
}
