package repository

import (
	"context"
	"database/sql"
	"sync"

	"github.com/google/uuid"
)

// MemoryUnitsRepo: 用于 DB 未就绪时的联测（UnitList.vue 需要 buildings -> units -> rooms/beds 的最小闭环）
// - 按 tenant_id 隔离
// - IDs 使用 uuid
// - 仅保证与 owlFront 当前调用形态兼容（不做复杂校验/唯一约束）
type MemoryUnitsRepo struct {
	mu sync.RWMutex

	// buildings keyed by tenant
	buildings map[string]map[string]map[string]any // tenantID -> buildingID -> building json

	// units keyed by tenant
	units map[string]map[string]Unit // tenantID -> unitID -> Unit

	// rooms keyed by unitID
	rooms map[string]map[string]Room // unitID -> roomID -> Room

	// beds keyed by roomID
	beds map[string]map[string]Bed // roomID -> bedID -> Bed
}

func NewMemoryUnitsRepo() *MemoryUnitsRepo {
	return &MemoryUnitsRepo{
		buildings: map[string]map[string]map[string]any{},
		units:     map[string]map[string]Unit{},
		rooms:     map[string]map[string]Room{},
		beds:      map[string]map[string]Bed{},
	}
}

// ---- buildings (extra methods used by HTTP handler via type assertion) ----

func (r *MemoryUnitsRepo) CreateBuilding(_ context.Context, tenantID string, payload map[string]any) (map[string]any, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tenantID == "" {
		return map[string]any{}, nil
	}
	if r.buildings[tenantID] == nil {
		r.buildings[tenantID] = map[string]map[string]any{}
	}

	buildingName, _ := payload["building_name"].(string)
	if buildingName == "" {
		buildingName = "-"
	}
	branchTag, _ := payload["branch_tag"].(string)
	if branchTag == "" {
		branchTag = "-"
	}
	floors := 1
	if v, ok := payload["floors"].(float64); ok && int(v) > 0 {
		floors = int(v)
	}
	if v, ok := payload["floors"].(int); ok && v > 0 {
		floors = v
	}

	id := uuid.NewString()
	b := map[string]any{
		"building_id":   id,
		"building_name": buildingName,
		"floors":        floors,
		"tenant_id":     tenantID,
		"branch_tag":    branchTag,
	}
	r.buildings[tenantID][id] = b
	return b, nil
}

func (r *MemoryUnitsRepo) UpdateBuilding(_ context.Context, tenantID, buildingID string, payload map[string]any) (map[string]any, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tenantID == "" || buildingID == "" {
		return map[string]any{}, nil
	}
	if r.buildings[tenantID] == nil {
		r.buildings[tenantID] = map[string]map[string]any{}
	}
	b, ok := r.buildings[tenantID][buildingID]
	if !ok {
		// create-on-update for dev convenience
		b = map[string]any{
			"building_id":   buildingID,
			"building_name": "-",
			"floors":        1,
			"tenant_id":     tenantID,
			"branch_tag":    "-",
		}
	}
	if v, ok := payload["building_name"].(string); ok && v != "" {
		b["building_name"] = v
	}
	if v, ok := payload["branch_tag"].(string); ok && v != "" {
		b["branch_tag"] = v
	}
	if v, ok := payload["floors"].(float64); ok && int(v) > 0 {
		b["floors"] = int(v)
	}
	if v, ok := payload["floors"].(int); ok && v > 0 {
		b["floors"] = v
	}
	r.buildings[tenantID][buildingID] = b
	return b, nil
}

func (r *MemoryUnitsRepo) DeleteBuilding(_ context.Context, tenantID, buildingID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if tenantID == "" || buildingID == "" {
		return nil
	}
	if r.buildings[tenantID] != nil {
		delete(r.buildings[tenantID], buildingID)
	}
	return nil
}

// ---- UnitsRepo interface ----

func (r *MemoryUnitsRepo) ListBuildings(_ context.Context, tenantID string, branchTag string) ([]map[string]any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []map[string]any{}
	if tenantID == "" {
		return out, nil
	}
	for _, b := range r.buildings[tenantID] {
		if branchTag != "" {
			if lt, _ := b["branch_tag"].(string); lt != branchTag {
				continue
			}
		}
		out = append(out, b)
	}
	return out, nil
}

func (r *MemoryUnitsRepo) ListUnits(_ context.Context, tenantID string, filters map[string]string, page, size int) ([]Unit, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if tenantID == "" {
		return []Unit{}, 0, nil
	}
	all := make([]Unit, 0, len(r.units[tenantID]))
	for _, u := range r.units[tenantID] {
		if v := filters["building"]; v != "" && u.Building != v {
			continue
		}
		if v := filters["floor"]; v != "" && u.Floor != v {
			continue
		}
		if v := filters["branch_tag"]; v != "" && u.BranchTag != v {
			continue
		}
		all = append(all, u)
	}
	total := len(all)
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 100
	}
	start := (page - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func (r *MemoryUnitsRepo) GetUnit(_ context.Context, tenantID, unitID string) (*Unit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if tenantID == "" || unitID == "" {
		return nil, sql.ErrNoRows
	}
	u, ok := r.units[tenantID][unitID]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return &u, nil
}

func (r *MemoryUnitsRepo) CreateUnit(_ context.Context, tenantID string, payload map[string]any) (*Unit, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if tenantID == "" {
		return nil, nil
	}
	if r.units[tenantID] == nil {
		r.units[tenantID] = map[string]Unit{}
	}

	unitName, _ := payload["unit_name"].(string)
	unitNumber, _ := payload["unit_number"].(string)
	unitType, _ := payload["unit_type"].(string)
	if unitType == "" {
		unitType = "Facility"
	}
	building, _ := payload["building"].(string)
	if building == "" {
		building = "-"
	}
	floor, _ := payload["floor"].(string)
	if floor == "" {
		floor = "1F"
	}
	branchTag, _ := payload["branch_tag"].(string)
	if branchTag == "" {
		branchTag = "-"
	}
	timezone, _ := payload["timezone"].(string)
	if timezone == "" {
		timezone = "America/Denver"
	}

	id := uuid.NewString()
	u := Unit{
		UnitID:     id,
		TenantID:   tenantID,
		BranchTag:  branchTag,
		UnitName:   unitName,
		Building:   building,
		Floor:      floor,
		UnitNumber: unitNumber,
		UnitType:   unitType,
		Timezone:   timezone,
	}
	r.units[tenantID][id] = u
	return &u, nil
}

func (r *MemoryUnitsRepo) UpdateUnit(_ context.Context, tenantID, unitID string, payload map[string]any) (*Unit, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if tenantID == "" || unitID == "" {
		return nil, sql.ErrNoRows
	}
	u, ok := r.units[tenantID][unitID]
	if !ok {
		return nil, sql.ErrNoRows
	}
	if v, ok := payload["unit_name"].(string); ok && v != "" {
		u.UnitName = v
	}
	if v, ok := payload["unit_number"].(string); ok && v != "" {
		u.UnitNumber = v
	}
	if v, ok := payload["unit_type"].(string); ok && v != "" {
		u.UnitType = v
	}
	if v, ok := payload["timezone"].(string); ok && v != "" {
		u.Timezone = v
	}
	r.units[tenantID][unitID] = u
	return &u, nil
}

func (r *MemoryUnitsRepo) DeleteUnit(_ context.Context, tenantID, unitID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if tenantID == "" || unitID == "" {
		return nil
	}
	delete(r.units[tenantID], unitID)
	delete(r.rooms, unitID)
	return nil
}

func (r *MemoryUnitsRepo) ListRoomsWithBeds(_ context.Context, unitID string) ([]map[string]any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []map[string]any{}
	for _, room := range r.rooms[unitID] {
		rm := room.ToJSON()
		// attach beds
		bs := []any{}
		for _, bed := range r.beds[room.RoomID] {
			bs = append(bs, bed.ToJSON())
		}
		rm["beds"] = bs
		out = append(out, rm)
	}
	return out, nil
}

func (r *MemoryUnitsRepo) CreateRoom(_ context.Context, unitID string, payload map[string]any) (*Room, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if unitID == "" {
		return nil, nil
	}
	if r.rooms[unitID] == nil {
		r.rooms[unitID] = map[string]Room{}
	}
	roomName, _ := payload["room_name"].(string)
	if roomName == "" {
		roomName = "Room"
	}
	id := uuid.NewString()
	room := Room{
		RoomID:   id,
		UnitID:   unitID,
		RoomName: roomName,
	}
	r.rooms[unitID][id] = room
	return &room, nil
}

func (r *MemoryUnitsRepo) UpdateRoom(_ context.Context, roomID string, payload map[string]any) (*Room, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for unitID, rms := range r.rooms {
		if room, ok := rms[roomID]; ok {
			if v, ok := payload["room_name"].(string); ok && v != "" {
				room.RoomName = v
			}
			r.rooms[unitID][roomID] = room
			return &room, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *MemoryUnitsRepo) DeleteRoom(_ context.Context, roomID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for unitID := range r.rooms {
		if r.rooms[unitID] != nil {
			delete(r.rooms[unitID], roomID)
		}
	}
	delete(r.beds, roomID)
	return nil
}

func (r *MemoryUnitsRepo) ListBeds(_ context.Context, roomID string) ([]Bed, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []Bed{}
	for _, b := range r.beds[roomID] {
		out = append(out, b)
	}
	return out, nil
}

func (r *MemoryUnitsRepo) CreateBed(_ context.Context, roomID string, payload map[string]any) (*Bed, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if roomID == "" {
		return nil, nil
	}
	if r.beds[roomID] == nil {
		r.beds[roomID] = map[string]Bed{}
	}
	bedName, _ := payload["bed_name"].(string)
	if bedName == "" {
		bedName = "Bed"
	}
	id := uuid.NewString()
	b := Bed{
		BedID:   id,
		RoomID:  roomID,
		BedName: bedName,
	}
	r.beds[roomID][id] = b
	return &b, nil
}

func (r *MemoryUnitsRepo) UpdateBed(_ context.Context, bedID string, payload map[string]any) (*Bed, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for roomID, beds := range r.beds {
		if b, ok := beds[bedID]; ok {
			if v, ok := payload["bed_name"].(string); ok && v != "" {
				b.BedName = v
			}
			r.beds[roomID][bedID] = b
			return &b, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *MemoryUnitsRepo) DeleteBed(_ context.Context, bedID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for roomID := range r.beds {
		if r.beds[roomID] != nil {
			delete(r.beds[roomID], bedID)
		}
	}
	return nil
}
