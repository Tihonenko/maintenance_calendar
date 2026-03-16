
CREATE TABLE vehicles (
    id SERIAL PRIMARY KEY,
    vin VARCHAR(50) UNIQUE NOT NULL,
    total_mileage DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_engine_hours DOUBLE PRECISION NOT NULL DEFAULT 0,
    avg_speed DOUBLE PRECISION NOT NULL CHECK (avg_speed BETWEEN 10.5 AND 16.3),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE maintenance_types (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    interval_km INT,
    interval_hours INT,
    is_cascading BOOLEAN DEFAULT FALSE,
    is_one_time BOOLEAN DEFAULT FALSE,
    is_seasonal BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE maintenance_actions (
    id SERIAL PRIMARY KEY,
    type_id INT NOT NULL REFERENCES maintenance_types(id) ON DELETE CASCADE,
    system_node VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE service_records (
    id SERIAL PRIMARY KEY,
    vehicle_id INT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    type_id INT NOT NULL REFERENCES maintenance_types(id),
    status VARCHAR(20) NOT NULL DEFAULT 'PLANNED',
    calculated_date DATE NOT NULL,
    scheduled_date DATE NOT NULL,
    completion_date DATE,
    mileage_at_completion DOUBLE PRECISION,
    hours_at_completion DOUBLE PRECISION,
    is_rescheduled BOOLEAN DEFAULT FALSE,
    ui_status VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_vehicles_vin ON vehicles(vin);
CREATE INDEX idx_maintenance_types_code ON maintenance_types(code);
CREATE INDEX idx_maintenance_actions_type_id ON maintenance_actions(type_id);
CREATE INDEX idx_service_records_vehicle_id ON service_records(vehicle_id);
CREATE INDEX idx_service_records_status ON service_records(status);
CREATE INDEX idx_service_records_dates ON service_records(calculated_date, scheduled_date);

CREATE TABLE IF NOT EXISTS service_record_items (
    id SERIAL PRIMARY KEY,
    service_record_id INT NOT NULL REFERENCES service_records(id) ON DELETE CASCADE,
    action_id INT NOT NULL REFERENCES maintenance_actions(id) ON DELETE RESTRICT,
    is_passed BOOLEAN NOT NULL DEFAULT TRUE,
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_service_record_items_record_id ON service_record_items(service_record_id);

ALTER TABLE service_records 
    DROP COLUMN IF EXISTS ui_status;

UPDATE service_records 
SET status = 'PLANNED' 
WHERE status NOT IN ('PLANNED', 'IN_PROGRESS', 'DONE', 'CANCELLED', 'OVERDUE');

ALTER TABLE service_records 
    ADD CONSTRAINT chk_status 
    CHECK (status IN ('PLANNED', 'IN_PROGRESS', 'DONE', 'CANCELLED', 'OVERDUE'));
