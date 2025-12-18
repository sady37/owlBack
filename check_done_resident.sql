-- Check done resident's profile, PHI, and contacts
-- First, find the resident by nickname or email

-- 1. Find resident_id by nickname
SELECT 
    r.resident_id::text,
    r.nickname,
    r.resident_account,
    r.status,
    r.service_level,
    r.admission_date,
    r.discharge_date,
    r.family_tag,
    r.can_view_status,
    r.note,
    r.unit_id::text,
    r.room_id::text,
    r.bed_id::text,
    CASE WHEN r.phone_hash IS NOT NULL THEN encode(r.phone_hash, 'hex') ELSE NULL END as phone_hash_hex,
    CASE WHEN r.email_hash IS NOT NULL THEN encode(r.email_hash, 'hex') ELSE NULL END as email_hash_hex
FROM residents r
WHERE LOWER(r.nickname) = 'done'
LIMIT 1;

-- 2. Get PHI data for done
SELECT 
    rp.phi_id::text,
    rp.first_name,
    rp.last_name,
    rp.gender,
    rp.date_of_birth,
    rp.resident_phone,
    rp.resident_email,
    rp.weight_lb,
    rp.height_ft,
    rp.height_in,
    rp.mobility_level,
    rp.tremor_status,
    rp.mobility_aid,
    rp.adl_assistance,
    rp.comm_status,
    rp.has_hypertension,
    rp.has_hyperlipaemia,
    rp.has_hyperglycaemia,
    rp.has_stroke_history,
    rp.has_paralysis,
    rp.has_alzheimer,
    rp.medical_history,
    rp.HIS_resident_name,
    rp.HIS_resident_admission_date,
    rp.HIS_resident_discharge_date,
    rp.home_address_street,
    rp.home_address_city,
    rp.home_address_state,
    rp.home_address_postal_code,
    rp.plus_code
FROM residents r
JOIN resident_phi rp ON r.resident_id = rp.resident_id
WHERE LOWER(r.nickname) = 'done';

-- 3. Get contacts for done
SELECT 
    rc.contact_id::text,
    rc.slot,
    rc.contact_family_tag,
    rc.is_enabled,
    rc.relationship,
    rc.is_emergency_contact,
    rc.alert_time_window,
    rc.contact_first_name,
    rc.contact_last_name,
    rc.contact_phone,
    rc.contact_email,
    rc.receive_sms,
    rc.receive_email,
    CASE WHEN rc.phone_hash IS NOT NULL THEN encode(rc.phone_hash, 'hex') ELSE NULL END as phone_hash_hex,
    CASE WHEN rc.email_hash IS NOT NULL THEN encode(rc.email_hash, 'hex') ELSE NULL END as email_hash_hex,
    CASE WHEN rc.password_hash IS NOT NULL THEN '***' ELSE NULL END as password_hash_exists
FROM residents r
JOIN resident_contacts rc ON r.resident_id = rc.resident_id
WHERE LOWER(r.nickname) = 'done'
ORDER BY rc.slot;

