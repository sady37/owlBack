package roomutil

import "testing"

func TestClassifyRoomType(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Bathroom", RoomTypeBathroom},
		{"BATHROOM 1F", RoomTypeBathroom},
		{"WC", RoomTypeBathroom},
		{"wc1F", RoomTypeBathroom},
		{"Restroom A", RoomTypeBathroom},
		{"Toilet", RoomTypeBathroom},
		{"Master Bedroom", RoomTypeBedroom},
		{"bed room 2", RoomTypeBedroom},
		{"Kitchen", RoomTypeKitchen},
		{"Living Room", RoomTypeOther},
		{"", RoomTypeOther},
		{"Garage", RoomTypeOther},
	}
	for _, c := range cases {
		if got := ClassifyRoomType(c.in); got != c.want {
			t.Errorf("ClassifyRoomType(%q) = %q, want %q", c.in, got, c.want)
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
