-- seed_demo_data.sql
-- 在 owl_v2 demo tenant (slot=3) + system tenant (slot=1) 下生成完整测试数据。
-- 幂等：所有 INSERT 用 ON CONFLICT DO NOTHING；多次运行结果一致。
--
-- 数据规模：
--   demo tenant：4 branches / 4 sites / 9 units / 17 rooms / 9 beds / 10 residents / 12 users
--   system tenant：1 sysadmin
--
-- 跑法：bash scripts/seed_demo_data.sh

-- =========================================================================
-- BRANCHES (4 in demo, branch_slot 1..4)
-- =========================================================================
INSERT INTO branches (spatial_prefix, branch_slot, branch_name, timezone) VALUES
    ('fd00:0:3:100::/56', 1, 'Denver',   'America/Denver'),
    ('fd00:0:3:200::/56', 2, 'Spring',   'America/Denver'),     -- Spring, CO (Denver 旁边)
    ('fd00:0:3:300::/56', 3, 'SanDiego', 'America/Los_Angeles'),
    ('fd00:0:3:400::/56', 4, 'ShenZhen', 'Asia/Shanghai')
ON CONFLICT (spatial_prefix) DO NOTHING;

-- =========================================================================
-- SITES (one per branch, building 0 floor 1, site_slot=0x01)
-- =========================================================================
INSERT INTO sites (spatial_prefix, site_slot, building, floor, site_name) VALUES
    ('fd00:0:3:101::/64', 1, 0, 1, 'Denver Bldg-0 Floor-1'),
    ('fd00:0:3:201::/64', 1, 0, 1, 'Spring Bldg-0 Floor-1'),
    ('fd00:0:3:301::/64', 1, 0, 1, 'SanDiego Bldg-0 Floor-1'),
    ('fd00:0:3:401::/64', 1, 0, 1, 'ShenZhen Bldg-0 Floor-1')
ON CONFLICT (spatial_prefix) DO NOTHING;

-- =========================================================================
-- UNITS (9 total)
-- =========================================================================
INSERT INTO units (spatial_prefix, unit_slot, unit_name, unit_type, unit_layout_type, is_public, is_shared_unit) VALUES
    -- Denver
    ('fd00:0:3:101:1::/80',   1,   'Denver Common',      'public',      'public_area', TRUE,  FALSE),
    ('fd00:0:3:101:65::/80',  101, 'Denver 101 VIP',     'residential', '2br_bath',    FALSE, FALSE),
    ('fd00:0:3:101:c9::/80',  201, 'Denver 201',         'residential', '1br_bath',    FALSE, FALSE),
    ('fd00:0:3:101:ca::/80',  202, 'Denver 202',         'residential', 'shared_dorm', FALSE, TRUE),
    -- Spring
    ('fd00:0:3:201:cc::/80',  204, 'Spring 204',         'residential', '1br_bath',    FALSE, FALSE),
    -- SanDiego
    ('fd00:0:3:301:1::/80',   1,   'SanDiego BlueOcean', 'public',      'public_area', TRUE,  FALSE),
    -- ShenZhen
    ('fd00:0:3:401:65::/80',  101, 'ShenZhen 101',       'residential', '1br_bath',    FALSE, FALSE),
    ('fd00:0:3:401:66::/80',  102, 'ShenZhen 102',       'residential', '1br_bath',    FALSE, FALSE),
    ('fd00:0:3:401:67::/80',  103, 'ShenZhen 103',       'residential', '1br_bath',    FALSE, FALSE)
ON CONFLICT (spatial_prefix) DO NOTHING;

-- =========================================================================
-- ROOMS (17 total)
-- =========================================================================
INSERT INTO rooms (spatial_prefix, room_slot, room_name, room_type, is_primary) VALUES
    -- Denver Common (3 public rooms; no primary)
    ('fd00:0:3:101:1:100::/88',  1, 'LivingRoom',  'livingroom', FALSE),
    ('fd00:0:3:101:1:200::/88',  2, 'ReadingRoom', 'other',      FALSE),
    ('fd00:0:3:101:1:300::/88',  3, 'Kitchen',     'kitchen',    FALSE),
    -- Denver 101 VIP (Room1=primary bedroom)
    ('fd00:0:3:101:65:100::/88', 1, 'Room1',       'bedroom',    TRUE),
    ('fd00:0:3:101:65:200::/88', 2, 'Room2',       'bedroom',    FALSE),
    ('fd00:0:3:101:65:300::/88', 3, 'Bathroom',    'bathroom',   FALSE),
    -- Denver 201 (single apartment)
    ('fd00:0:3:101:c9:100::/88', 1, 'Bedroom',     'bedroom',    TRUE),
    ('fd00:0:3:101:c9:200::/88', 2, 'Bathroom',    'bathroom',   FALSE),
    -- Denver 202 (shared dorm — 1 bedroom, 2 beds)
    ('fd00:0:3:101:ca:100::/88', 1, 'Bedroom',     'bedroom',    TRUE),
    -- Spring 204
    ('fd00:0:3:201:cc:100::/88', 1, 'Bedroom',     'bedroom',    TRUE),
    ('fd00:0:3:201:cc:200::/88', 2, 'Bathroom',    'bathroom',   FALSE),
    -- SanDiego BlueOcean (1 main hall public)
    ('fd00:0:3:301:1:100::/88',  1, 'MainHall',    'lobby',      FALSE),
    -- ShenZhen 101/102/103 (each: bedroom primary + bathroom)
    ('fd00:0:3:401:65:100::/88', 1, 'Bedroom',     'bedroom',    TRUE),
    ('fd00:0:3:401:65:200::/88', 2, 'Bathroom',    'bathroom',   FALSE),
    ('fd00:0:3:401:66:100::/88', 1, 'Bedroom',     'bedroom',    TRUE),
    ('fd00:0:3:401:66:200::/88', 2, 'Bathroom',    'bathroom',   FALSE),
    ('fd00:0:3:401:67:100::/88', 1, 'Bedroom',     'bedroom',    TRUE),
    ('fd00:0:3:401:67:200::/88', 2, 'Bathroom',    'bathroom',   FALSE)
ON CONFLICT (spatial_prefix) DO NOTHING;

-- =========================================================================
-- BEDS (9 total)
-- =========================================================================
INSERT INTO beds (spatial_prefix, bed_slot, bed_name) VALUES
    -- Denver 101 VIP
    ('fd00:0:3:101:65:101::/96', 1, 'Room1-BedA'),
    ('fd00:0:3:101:65:201::/96', 1, 'Room2-BedA'),
    -- Denver 201
    ('fd00:0:3:101:c9:101::/96', 1, 'BedA'),
    -- Denver 202 shared
    ('fd00:0:3:101:ca:101::/96', 1, 'BedA'),
    ('fd00:0:3:101:ca:102::/96', 2, 'BedB'),
    -- Spring 204
    ('fd00:0:3:201:cc:101::/96', 1, 'BedA'),
    -- ShenZhen 101/102/103
    ('fd00:0:3:401:65:101::/96', 1, 'BedA'),
    ('fd00:0:3:401:66:101::/96', 1, 'BedA'),
    ('fd00:0:3:401:67:101::/96', 1, 'BedA')
ON CONFLICT (spatial_prefix) DO NOTHING;

-- =========================================================================
-- RESIDENTS (10：7 US + 3 CN；slots 1..10)
-- 注：phone/email 暂存 notes 字段（明文 placeholder，TEST DATA）；
--    生产 PHI 须经 KMS 加密入 resident_phi 表。
-- =========================================================================
INSERT INTO residents (hoa, resident_id, resident_slot, nickname, gender, birth_year, move_in_date, status, service_tier, notes) VALUES
    ('fd00:0:3:ff01:1::/128',  gen_random_uuid(),  1, 'John',     'M', 1942, '2025-03-15', 'active', 'premium',
        '[TEST] phone: +1-303-555-0101 | email: john.smith@example.com'),
    ('fd00:0:3:ff01:2::/128',  gen_random_uuid(),  2, 'Mary',     'F', 1947, '2025-05-01', 'active', 'standard',
        '[TEST] phone: +1-303-555-0102 | email: mary.johnson@example.com'),
    ('fd00:0:3:ff01:3::/128',  gen_random_uuid(),  3, 'Robert',   'M', 1950, '2025-08-10', 'active', 'standard',
        '[TEST] phone: +1-303-555-0103 | email: robert.brown@example.com'),
    ('fd00:0:3:ff01:4::/128',  gen_random_uuid(),  4, 'Patricia', 'F', 1953, '2025-09-22', 'active', 'standard',
        '[TEST] phone: +1-303-555-0104 | email: patricia.davis@example.com'),
    ('fd00:0:3:ff01:5::/128',  gen_random_uuid(),  5, 'DanFa',    'M', 1945, '2025-04-30', 'active', 'standard',
        '[TEST] phone: +1-281-555-0105 | email: danfa@example.com'),
    ('fd00:0:3:ff01:6::/128',  gen_random_uuid(),  6, 'William',  'M', 1958, NULL,         'active', NULL,
        '[TEST] phone: +1-303-555-0106 | email: william.anderson@example.com | unassigned, on waiting list'),
    ('fd00:0:3:ff01:7::/128',  gen_random_uuid(),  7, 'Linda',    'F', 1955, NULL,         'active', NULL,
        '[TEST] phone: +1-303-555-0107 | email: linda.wilson@example.com | unassigned, on waiting list'),
    ('fd00:0:3:ff01:8::/128',  gen_random_uuid(),  8, 'MoM',      'F', 1943, '2024-09-01', 'active', 'premium',
        '[TEST] phone: +86-138-1234-5678 | email: mom@example.cn'),
    ('fd00:0:3:ff01:9::/128',  gen_random_uuid(),  9, 'Frand',    'M', 1948, '2024-11-15', 'active', 'standard',
        '[TEST] phone: +86-138-2345-6789 | email: frand@example.cn'),
    ('fd00:0:3:ff01:a::/128',  gen_random_uuid(), 10, 'Ton',      'M', 1956, '2025-02-01', 'active', 'standard',
        '[TEST] phone: +86-138-3456-7890 | email: ton@example.cn')
ON CONFLICT (hoa) DO NOTHING;

-- =========================================================================
-- RESIDENT_UNIT (8 active assignments；2 unassigned 备选)
-- =========================================================================
INSERT INTO resident_unit (resident_hoa, spatial_prefix, valid_from, move_reason)
SELECT * FROM (VALUES
    ('fd00:0:3:ff01:1::/128'::INET,  'fd00:0:3:101:65::/80'::INET,     '2025-03-15'::TIMESTAMPTZ, 'initial'),  -- John VIP unit /80
    ('fd00:0:3:ff01:2::/128'::INET,  'fd00:0:3:101:c9::/80'::INET,     '2025-05-01'::TIMESTAMPTZ, 'initial'),  -- Mary single unit /80
    ('fd00:0:3:ff01:3::/128'::INET,  'fd00:0:3:101:ca:101::/96'::INET, '2025-08-10'::TIMESTAMPTZ, 'initial'),  -- Robert bed
    ('fd00:0:3:ff01:4::/128'::INET,  'fd00:0:3:101:ca:102::/96'::INET, '2025-09-22'::TIMESTAMPTZ, 'initial'),  -- Patricia bed
    -- 单人独享 1br_bath unit 用 /80 unit 绑定（与 Mary 同模式；浴室 radar alarm 也能 route 到该 resident）
    ('fd00:0:3:ff01:5::/128'::INET,  'fd00:0:3:201:cc::/80'::INET,     '2025-04-30'::TIMESTAMPTZ, 'initial'),  -- DanFa unit
    ('fd00:0:3:ff01:8::/128'::INET,  'fd00:0:3:401:65::/80'::INET,     '2024-09-01'::TIMESTAMPTZ, 'initial'),  -- MoM unit
    ('fd00:0:3:ff01:9::/128'::INET,  'fd00:0:3:401:66::/80'::INET,     '2024-11-15'::TIMESTAMPTZ, 'initial'),  -- Frand unit
    ('fd00:0:3:ff01:a::/128'::INET,  'fd00:0:3:401:67::/80'::INET,     '2025-02-01'::TIMESTAMPTZ, 'initial')   -- Ton unit
) AS v(resident_hoa, spatial_prefix, valid_from, move_reason)
WHERE NOT EXISTS (
    SELECT 1 FROM resident_unit ru
    WHERE ru.resident_hoa = v.resident_hoa AND ru.valid_to IS NULL
);

-- =========================================================================
-- USERS - system tenant (1 sysadmin)
-- =========================================================================
INSERT INTO users (tenant_prefix, username, password_hash, mobile_pin_hash, nickname, full_name, email, status, notify_mode) VALUES
    ('fd00:0:1::/48', 'sysadmin', crypt('ChangeMe@123', gen_salt('bf')), crypt('1212', gen_salt('bf')), 'sysadmin', 'System Admin', 'sysadmin@owl.internal', 'active', 'forever')
ON CONFLICT (username) DO NOTHING;

-- =========================================================================
-- USERS - demo tenant：3 admin (no hoa) + 5 caregiver + 4 nurse (with hoa)
-- =========================================================================
INSERT INTO users (tenant_prefix, username, password_hash, mobile_pin_hash, nickname, full_name, email, status, notify_mode) VALUES
    ('fd00:0:3::/48', 'admin', crypt('Ts123@123',  gen_salt('bf')), crypt('1212', gen_salt('bf')), 'admin', 'Demo Admin', 'admin@wisefido.com', 'active', 'login_only'),
    ('fd00:0:3::/48', 'demo',  crypt('Demo@2026',  gen_salt('bf')), crypt('1212', gen_salt('bf')), 'demo',  'Demo User',  'demo@wisefido.com',  'active', 'login_only'),
    ('fd00:0:3::/48', 'hunzi', crypt('hunzi@2026', gen_salt('bf')), crypt('1212', gen_salt('bf')), 'hunzi', 'HunZi',      'hunzi@wisefido.com', 'active', 'login_only')
ON CONFLICT (username) DO NOTHING;

-- caregiver/nurse (hoa filled，每分支挂 1 个；Denver 多 1 caregiver)
INSERT INTO users (tenant_prefix, username, password_hash, mobile_pin_hash, nickname, full_name, email, hoa, subject_slot, employee_code, role, hire_date, status, notify_mode) VALUES
    ('fd00:0:3::/48', 'caregiver_denver_1',  crypt('Caregiver@2026', gen_salt('bf')), crypt('1212', gen_salt('bf')), 'Alice', 'Alice Brown',  'alice@wisefido.com',  'fd00:0:3:ff02:1::/128', 1, 'CG-DEN-001', 'caregiver', '2025-01-15', 'active', 'login_only'),
    ('fd00:0:3::/48', 'caregiver_denver_2',  crypt('Caregiver@2026', gen_salt('bf')), crypt('1212', gen_salt('bf')), 'Bob',   'Bob Carter',   'bob@wisefido.com',    'fd00:0:3:ff02:2::/128', 2, 'CG-DEN-002', 'caregiver', '2025-02-01', 'active', 'login_only'),
    ('fd00:0:3::/48', 'caregiver_spring',    crypt('Caregiver@2026', gen_salt('bf')), crypt('1212', gen_salt('bf')), 'Carol', 'Carol Davis',  'carol@wisefido.com',  'fd00:0:3:ff02:3::/128', 3, 'CG-SPR-001', 'caregiver', '2025-03-10', 'active', 'login_only'),
    ('fd00:0:3::/48', 'caregiver_sandiego',  crypt('Caregiver@2026', gen_salt('bf')), crypt('1212', gen_salt('bf')), 'Dave',  'Dave Edwards', 'dave@wisefido.com',   'fd00:0:3:ff02:4::/128', 4, 'CG-SDG-001', 'caregiver', '2025-04-05', 'active', 'login_only'),
    ('fd00:0:3::/48', 'caregiver_shenzhen',  crypt('Caregiver@2026', gen_salt('bf')), crypt('1212', gen_salt('bf')), 'Ling',  'Wang Ling',    'ling@wisefido.com',   'fd00:0:3:ff02:5::/128', 5, 'CG-SHE-001', 'caregiver', '2024-08-01', 'active', 'login_only'),
    ('fd00:0:3::/48', 'nurse_denver',        crypt('Nurse@2026',     gen_salt('bf')), crypt('1212', gen_salt('bf')), 'Ema',   'Ema Foster',   'ema@wisefido.com',    'fd00:0:3:ff02:6::/128', 6, 'RN-DEN-001', 'nurse',     '2024-06-01', 'active', 'login_only'),
    ('fd00:0:3::/48', 'nurse_spring',        crypt('Nurse@2026',     gen_salt('bf')), crypt('1212', gen_salt('bf')), 'Fred',  'Fred Garcia',  'fred@wisefido.com',   'fd00:0:3:ff02:7::/128', 7, 'RN-SPR-001', 'nurse',     '2024-07-01', 'active', 'login_only'),
    ('fd00:0:3::/48', 'nurse_sandiego',      crypt('Nurse@2026',     gen_salt('bf')), crypt('1212', gen_salt('bf')), 'Grace', 'Grace Harris', 'grace@wisefido.com',  'fd00:0:3:ff02:8::/128', 8, 'RN-SDG-001', 'nurse',     '2024-08-15', 'active', 'login_only'),
    ('fd00:0:3::/48', 'nurse_shenzhen',      crypt('Nurse@2026',     gen_salt('bf')), crypt('1212', gen_salt('bf')), 'Hong',  'Liu Hong',     'hong@wisefido.com',   'fd00:0:3:ff02:9::/128', 9, 'RN-SHE-001', 'nurse',     '2024-05-15', 'active', 'login_only')
ON CONFLICT (username) DO NOTHING;

-- =========================================================================
-- USER_ROLES：assign roles by branch /56 scope
--   sysadmin       → platform_admin (全局 NULL scope)
--   admin          → tenant_admin   (tenant /48 = NULL)
--   demo / hunzi   → manager        (tenant /48)
--   caregiver/nurse→ caregiver/nurse(branch /56)
-- =========================================================================
INSERT INTO user_roles (user_id, role_id, scope)
SELECT u.user_id, r.role_id, v.scope
FROM (VALUES
    ('sysadmin',           'platform_admin', NULL::INET),
    ('admin',              'tenant_admin',   NULL),
    ('demo',               'manager',        NULL),
    ('hunzi',              'manager',        NULL),
    ('caregiver_denver_1', 'caregiver', 'fd00:0:3:100::/56'::INET),
    ('caregiver_denver_2', 'caregiver', 'fd00:0:3:100::/56'),
    ('caregiver_spring',   'caregiver', 'fd00:0:3:200::/56'),
    ('caregiver_sandiego', 'caregiver', 'fd00:0:3:300::/56'),
    ('caregiver_shenzhen', 'caregiver', 'fd00:0:3:400::/56'),
    ('nurse_denver',       'nurse',     'fd00:0:3:100::/56'),
    ('nurse_spring',       'nurse',     'fd00:0:3:200::/56'),
    ('nurse_sandiego',     'nurse',     'fd00:0:3:300::/56'),
    ('nurse_shenzhen',     'nurse',     'fd00:0:3:400::/56')
) AS v(username, role_code, scope)
JOIN users u ON u.username = v.username
JOIN roles r ON r.role_code = v.role_code AND r.tenant_prefix IS NULL
ON CONFLICT (user_id, role_id, scope) DO NOTHING;

-- =========================================================================
-- DEVICE SPATIAL BINDING (19 devices: 11 HC2 radars + 8 BM8701-2 sleepads)
--
-- 绑定层级：
--   /96 bed prefix       — sleepad (床上) + bed-scoped radar
--   /88 room prefix      — room-level radar (浴室 / 公共空间 / 卧室通用)
--
-- /128 spatial_addr 派生：
--   bed-bound : bed_prefix(/96) + 16 bit zero (group7) + 16 bit dev_seq → fd00:...:101:0:1/128
--   room-bound: room_prefix(/88) + 8 bit byte11=0x00 (no bed) + 32 bit dev_seq → fd00:...:300::1/128
--
-- 设备从 device_factory_meta 池里按 device_uid 顺序取（前 11 HC2 + 前 8 BM8701-2）。
--
-- 注：ShenZhen-102 device label 'Hunzi_*' 与现 resident 'Frand' 不一致；
-- 此为 device-side label，不影响 resident_unit 绑定（用户后续可决定是否改 102 resident）。
-- =========================================================================
WITH
selected_radars AS (
    SELECT device_id, ROW_NUMBER() OVER (ORDER BY device_uid) AS n
    FROM device_factory_meta WHERE device_model = 'HC2' LIMIT 11
),
selected_sleepads AS (
    SELECT device_id, ROW_NUMBER() OVER (ORDER BY device_uid) AS n
    FROM device_factory_meta WHERE device_model = 'BM8701-2' LIMIT 8
),
radar_bindings (n, spatial_addr, label) AS (VALUES
    -- Denver
    (1::INT, 'fd00:0:3:101:65:101:0:1'::INET, '101_Bedroom_radar'),     -- 101 Room1.bedA
    (2,      'fd00:0:3:101:65:300::1'::INET,  '101_Bathroom_radar'),    -- 101 Bathroom (room-level)
    (3,      'fd00:0:3:101:1:100::1'::INET,   'LivingRoom_radar'),      -- public LivingRoom
    (4,      'fd00:0:3:101:1:200::1'::INET,   '101corn_radar'),         -- public ReadingRoom
    (5,      'fd00:0:3:101:1:300::1'::INET,   'Kitchen_radar'),         -- public Kitchen
    (6,      'fd00:0:3:101:c9:100::1'::INET,  '201_bedroom_radar'),     -- 201 Bedroom (room-level)
    (7,      'fd00:0:3:101:c9:200::1'::INET,  '201_bathroom_radar'),    -- 201 Bathroom (room-level)
    -- Spring
    (8,      'fd00:0:3:201:cc:101:0:1'::INET, 'DannaFa_Radar'),         -- 204 BedA radar
    -- SanDiego
    (9,      'fd00:0:3:301:1:100::1'::INET,   'San_Diego_Radar'),       -- BlueOcean MainHall
    -- ShenZhen
    (10,     'fd00:0:3:401:65:200::1'::INET,  'Mom_radar'),             -- ShenZhen 101 Bathroom
    (11,     'fd00:0:3:401:66:200::1'::INET,  'Hunzi_Radar')            -- ShenZhen 102 Bathroom
),
sleepad_bindings (n, spatial_addr, label) AS (VALUES
    -- Denver
    (1::INT, 'fd00:0:3:101:65:101:0:2'::INET, '101_Bedroom_sleepad'),   -- 101 Room1.bedA sleepad
    (2,      'fd00:0:3:101:65:201:0:1'::INET, 'GuestRoom_sleepad'),     -- 101 Room2.bedA sleepad
    (3,      'fd00:0:3:101:c9:101:0:1'::INET, '201_BeadA_sleepad'),     -- 201 BedA
    (4,      'fd00:0:3:101:ca:101:0:1'::INET, '202_sleepad'),           -- 202 BedA
    (5,      'fd00:0:3:101:ca:102:0:1'::INET, '203_sleepad'),           -- 202 BedB (label per user)
    -- Spring
    (6,      'fd00:0:3:201:cc:101:0:2'::INET, 'DannaFa_sleepad'),       -- 204 BedA sleepad
    -- ShenZhen
    (7,      'fd00:0:3:401:65:101:0:1'::INET, 'MoM_sleepad'),           -- ShenZhen 101 BedA
    (8,      'fd00:0:3:401:66:101:0:1'::INET, 'Hunzi_Sleepad')          -- ShenZhen 102 BedA
)
INSERT INTO devices (spatial_addr, device_id, monitoring_enabled)
SELECT b.spatial_addr, r.device_id, TRUE
FROM radar_bindings b JOIN selected_radars r ON r.n = b.n
UNION ALL
SELECT b.spatial_addr, s.device_id, TRUE
FROM sleepad_bindings b JOIN selected_sleepads s ON s.n = b.n
ON CONFLICT (spatial_addr) DO NOTHING;

-- 把 device label 写到 device_factory_meta.notes（追加到 import note 后）
UPDATE device_factory_meta dfm
SET notes = COALESCE(NULLIF(dfm.notes, ''), '') || ' | label=' || lbl.label
FROM (
    WITH
    selected_radars AS (
        SELECT device_id, ROW_NUMBER() OVER (ORDER BY device_uid) AS n
        FROM device_factory_meta WHERE device_model = 'HC2' LIMIT 11
    ),
    selected_sleepads AS (
        SELECT device_id, ROW_NUMBER() OVER (ORDER BY device_uid) AS n
        FROM device_factory_meta WHERE device_model = 'BM8701-2' LIMIT 8
    ),
    radar_bindings (n, label) AS (VALUES
        (1::INT,'101_Bedroom_radar'),(2,'101_Bathroom_radar'),(3,'LivingRoom_radar'),
        (4,'101corn_radar'),(5,'Kitchen_radar'),(6,'201_bedroom_radar'),
        (7,'201_bathroom_radar'),(8,'DannaFa_Radar'),(9,'San_Diego_Radar'),
        (10,'Mom_radar'),(11,'Hunzi_Radar')
    ),
    sleepad_bindings (n, label) AS (VALUES
        (1::INT,'101_Bedroom_sleepad'),(2,'GuestRoom_sleepad'),(3,'201_BeadA_sleepad'),
        (4,'202_sleepad'),(5,'203_sleepad'),(6,'DannaFa_sleepad'),
        (7,'MoM_sleepad'),(8,'Hunzi_Sleepad')
    )
    SELECT r.device_id, b.label FROM radar_bindings b JOIN selected_radars r ON r.n=b.n
    UNION ALL
    SELECT s.device_id, b.label FROM sleepad_bindings b JOIN selected_sleepads s ON s.n=b.n
) AS lbl
WHERE dfm.device_id = lbl.device_id
  AND COALESCE(dfm.notes, '') NOT LIKE '%label=%';

-- =========================================================================
-- 报告
-- =========================================================================
\echo '=== 完成统计 ==='
SELECT 'tenants'        AS scope, COUNT(*) AS n FROM tenants
UNION ALL SELECT 'branches',     COUNT(*) FROM branches
UNION ALL SELECT 'sites',        COUNT(*) FROM sites
UNION ALL SELECT 'units',        COUNT(*) FROM units
UNION ALL SELECT 'rooms',        COUNT(*) FROM rooms
UNION ALL SELECT 'beds',         COUNT(*) FROM beds
UNION ALL SELECT 'residents',    COUNT(*) FROM residents
UNION ALL SELECT 'resident_unit (active)', COUNT(*) FROM resident_unit WHERE valid_to IS NULL
UNION ALL SELECT 'users',        COUNT(*) FROM users
UNION ALL SELECT 'user_roles',   COUNT(*) FROM user_roles
UNION ALL SELECT 'devices (bound)', COUNT(*) FROM devices;
