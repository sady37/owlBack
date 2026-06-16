package roomengine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func LayoutHash(cfg RoomConfig) string {
	h := sha256.New()
	enc := json.NewEncoder(h)
	_ = enc.Encode(cfg.WallPolygon)
	_ = enc.Encode(cfg.Enters)
	_ = enc.Encode(cfg.EnterHeights)
	_ = enc.Encode(cfg.EnterTargets)
	_ = enc.Encode(cfg.RoomType)
	_ = enc.Encode(cfg.Beds)
	_ = enc.Encode(cfg.BedHeights)
	_ = enc.Encode(cfg.Toilets)
	_ = enc.Encode(cfg.ToiletHeights)
	_ = enc.Encode(cfg.Showers)
	_ = enc.Encode(cfg.ShowerHeights)
	_ = enc.Encode(cfg.Chairs)
	_ = enc.Encode(cfg.ChairHeights)
	_ = enc.Encode(cfg.Furnitures)
	_ = enc.Encode(cfg.FurnitureHeights)
	_ = enc.Encode(cfg.Interferes)
	_ = enc.Encode(cfg.InterfereHeights)
	_ = enc.Encode(cfg.Radar)
	return hex.EncodeToString(h.Sum(nil))
}
