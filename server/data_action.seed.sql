CREATE TEMP TABLE type_ids AS
SELECT 
    MAX(CASE WHEN code = 'AFTER_RUN' THEN id END) as after_run_id,
    MAX(CASE WHEN code = 'TO1' THEN id END) as to1_id,
    MAX(CASE WHEN code = 'TO2' THEN id END) as to2_id,
    MAX(CASE WHEN code = 'TO3' THEN id END) as to3_id,
    MAX(CASE WHEN code = 'DTO_50H' THEN id END) as dto_50h_id,
    MAX(CASE WHEN code = 'SEASONAL' THEN id END) as seasonal_id
FROM maintenance_types;

CREATE TEMP TABLE vehicle_to_schedule AS
SELECT 
    v.id as vehicle_id,
    v.vin,
    v.total_mileage,
    v.total_engine_hours,
    v.avg_speed
FROM vehicles v;


INSERT INTO service_records 
    (vehicle_id, type_id, status, calculated_date, scheduled_date, completion_date, mileage_at_completion, hours_at_completion, is_rescheduled, created_at, updated_at)
SELECT 
    vts.vehicle_id,
    t.to1_id,
    'DONE',
    CURRENT_DATE - INTERVAL '30 days',
    CURRENT_DATE - INTERVAL '30 days',
    CURRENT_DATE - INTERVAL '30 days',
    vts.total_mileage - 5000,
    vts.total_engine_hours - 250,
    FALSE,
    CURRENT_TIMESTAMP - INTERVAL '30 days',
    CURRENT_TIMESTAMP - INTERVAL '30 days'
FROM vehicle_to_schedule vts, type_ids t
WHERE vts.vehicle_id IN (1, 3, 5, 7, 9);


INSERT INTO service_records 
    (vehicle_id, type_id, status, calculated_date, scheduled_date, completion_date, mileage_at_completion, hours_at_completion, is_rescheduled, created_at, updated_at)
SELECT 
    vts.vehicle_id,
    t.to1_id,
    'PLANNED',
    CURRENT_DATE - INTERVAL '10 days',
    CURRENT_DATE - INTERVAL '10 days',
    NULL,
    NULL,
    NULL,
    FALSE,
    CURRENT_TIMESTAMP - INTERVAL '15 days',
    CURRENT_TIMESTAMP - INTERVAL '15 days'
FROM vehicle_to_schedule vts, type_ids t
WHERE vts.vehicle_id IN (2, 6);


INSERT INTO service_records 
    (vehicle_id, type_id, status, calculated_date, scheduled_date, completion_date, mileage_at_completion, hours_at_completion, is_rescheduled, created_at, updated_at)
SELECT 
    vts.vehicle_id,
    t.to1_id,
    'PLANNED',
    CURRENT_DATE - INTERVAL '2 days',
    CURRENT_DATE - INTERVAL '2 days',
    NULL,
    NULL,
    NULL,
    FALSE,
    CURRENT_TIMESTAMP - INTERVAL '5 days',
    CURRENT_TIMESTAMP - INTERVAL '5 days'
FROM vehicle_to_schedule vts, type_ids t
WHERE vts.vehicle_id IN (4, 8);


INSERT INTO service_records 
    (vehicle_id, type_id, status, calculated_date, scheduled_date, completion_date, mileage_at_completion, hours_at_completion, is_rescheduled, created_at, updated_at)
SELECT 
    vts.vehicle_id,
    t.to1_id,
    'PLANNED',
    CURRENT_DATE + INTERVAL '7 days',
    CURRENT_DATE + INTERVAL '7 days',
    NULL,
    NULL,
    NULL,
    FALSE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM vehicle_to_schedule vts, type_ids t
WHERE vts.vehicle_id IN (10);


INSERT INTO service_records 
    (vehicle_id, type_id, status, calculated_date, scheduled_date, completion_date, mileage_at_completion, hours_at_completion, is_rescheduled, created_at, updated_at)
SELECT 
    vts.vehicle_id,
    t.to2_id,
    'PLANNED',
    CURRENT_DATE + INTERVAL '20 days',
    CURRENT_DATE + INTERVAL '20 days',
    NULL,
    NULL,
    NULL,
    FALSE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM vehicle_to_schedule vts, type_ids t
WHERE vts.vehicle_id IN (1, 3, 5);

INSERT INTO service_records 
    (vehicle_id, type_id, status, calculated_date, scheduled_date, completion_date, mileage_at_completion, hours_at_completion, is_rescheduled, created_at, updated_at)
SELECT 
    vts.vehicle_id,
    t.to3_id,
    'PLANNED',
    CURRENT_DATE + INTERVAL '45 days',
    CURRENT_DATE + INTERVAL '45 days',
    NULL,
    NULL,
    NULL,
    FALSE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM vehicle_to_schedule vts, type_ids t
WHERE vts.vehicle_id IN (2, 4);


INSERT INTO service_records 
    (vehicle_id, type_id, status, calculated_date, scheduled_date, completion_date, mileage_at_completion, hours_at_completion, is_rescheduled, created_at, updated_at)
SELECT 
    vts.vehicle_id,
    t.seasonal_id,
    'PLANNED',
    CURRENT_DATE + INTERVAL '30 days',
    CURRENT_DATE + INTERVAL '30 days',
    NULL,
    NULL,
    NULL,
    FALSE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM vehicle_to_schedule vts, type_ids t
WHERE vts.vehicle_id IN (1, 2, 3);


INSERT INTO service_record_items (service_record_id, action_id, is_passed, comment, created_at)
SELECT 
    sr.id as service_record_id,
    ma.id as action_id,
    TRUE as is_passed,
    'Выполнено' as comment,
    CURRENT_TIMESTAMP - INTERVAL '30 days' as created_at
FROM service_records sr
JOIN maintenance_actions ma ON sr.type_id = ma.type_id
WHERE sr.status = 'DONE';


DROP TABLE IF EXISTS vehicle_to_schedule;
DROP TABLE IF EXISTS type_ids;
