SELECT * FROM service_records;

SELECT * from maintenance_types;

SELECT * FROM service_record_items;

SELECT * from maintenance_actions;

SELECT * from vehicles;

TRUNCATE TABLE 
    service_record_items,
    service_records,
    maintenance_actions,
    maintenance_types,
    vehicles
RESTART IDENTITY CASCADE;

INSERT INTO maintenance_types (code, name, interval_km, interval_hours, is_cascading, is_one_time, is_seasonal)
VALUES
    ('AFTER_RUN', 'Послеобкаточное ТО', 1000, 100, FALSE, TRUE, FALSE),
    ('TO1', 'ТО-1', 5000, 250, FALSE, FALSE, FALSE),
    ('TO2', 'ТО-2', 10000, 500, TRUE, FALSE, FALSE),
    ('TO3', 'ТО-3', 20000, 1000, TRUE, FALSE, FALSE);

-- ============================================
-- 2. Типы ТО - ДРУГИЕ ВИДЫ (ДТО)
-- ============================================
INSERT INTO maintenance_types (code, name, interval_km, interval_hours, is_cascading, is_one_time, is_seasonal)
VALUES
    ('DTO_50H', 'ДТО - 50ч', NULL, 50,  FALSE, FALSE, FALSE),
    ('DTO_2000H', 'ДТО - 2000ч', NULL, 2000, FALSE, FALSE, FALSE),
    ('DTO_2500H', 'ДТО - 2500ч', NULL, 2500,  FALSE, FALSE, FALSE),
    ('DTO_20000H', 'ДТО - 20000ч', NULL, 20000, FALSE, FALSE, FALSE),
    ('DTO_60000KM', 'ДТО - 60000км', 60000, NULL, FALSE, FALSE, FALSE),
    ('DTO_100000KM', 'ДТО - 100000км', 100000, NULL, FALSE, FALSE, FALSE),
    ('DTO_175000KM', 'ДТО - 175000км', 175000, NULL, FALSE, FALSE, FALSE),
    ('DTO_200000KM', 'ДТО - 200000км', 200000, NULL, FALSE, FALSE, FALSE),
    ('DTO_350000KM', 'ДТО - 350000км', 350000, NULL, FALSE, FALSE, FALSE),
    ('DTO_500000KM', 'ДТО - 500000км', 500000, NULL, FALSE, FALSE, FALSE);

-- ============================================
-- 3. Сезонное ТО
-- ============================================
INSERT INTO maintenance_types (code, name, interval_km, interval_hours, is_cascading, is_one_time, is_seasonal)
VALUES
    ('SEASONAL', 'Сезонное ТО', NULL, NULL, FALSE, FALSE, TRUE);

-- ============================================
-- 4. ДЕЙСТВИЯ ДЛЯ ТО1
-- ============================================
INSERT INTO maintenance_actions (type_id, system_node, description, sort_order)
SELECT id, system_node, description, ROW_NUMBER() OVER (ORDER BY system_node)
FROM (
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'TO1') as id,
        'Установка дизель генератора' as system_node,
        'Замена масла моторного' as description
    UNION ALL
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'TO1'),
        'Установка дизель генератора',
        'Замена масляных фильтров'
    UNION ALL
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'TO1'),
        'Установка дизель генератора',
        'Замена топливных фильтров'
    UNION ALL
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'TO1'),
        'Установка дизель генератора',
        'Выполнение диагностики двигателя'
    UNION ALL
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'TO1'),
        'Установка дизель генератора',
        'Очистка фильтроэлемента'
    UNION ALL
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'TO1'),
        'Дизель генератор',
        'Проверка работы подшипников'
    UNION ALL
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'TO1'),
        'Установка кабины',
        'Проверка всех электрических соединений'
) t;

-- ============================================
-- 5. ДЕЙСТВИЯ ДЛЯ ТО2 (только специфичные для ТО2)
-- ============================================
INSERT INTO maintenance_actions (type_id, system_node, description, sort_order)
SELECT id, system_node, description, ROW_NUMBER() OVER (ORDER BY system_node)
FROM (
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'TO2') as id,
        'Дизель генератор' as system_node,
        'Проверка состояния щеточного аппарата' as description
    UNION ALL
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'TO2'),
        'Дизель генератор',
        'Проверка биения контактных колец'
    UNION ALL
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'TO2'),
        'Установка кабины',
        'Осмотр привода компрессора'
    UNION ALL
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'TO2'),
        'Установка кабины',
        'Продувка сот конденсатора'
    UNION ALL
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'TO2'),
        'Установка кабины',
        'Промывка теплой водой конденсатора'
) t;

-- ============================================
-- 6. ДЕЙСТВИЯ ДЛЯ ТО3 (только специфичные для ТО3)
-- ============================================
INSERT INTO maintenance_actions (type_id, system_node, description, sort_order)
SELECT id, system_node, description, ROW_NUMBER() OVER (ORDER BY system_node)
FROM (
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'TO3') as id,
        'Дизель генератор' as system_node,
        'Продувка внутренних полостей генератора сухим сжатым воздухом' as description
    UNION ALL
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'TO3'),
        'Дизель генератор',
        'Проверка сопоставления изоляции обмоток генератора и траверсы'
    UNION ALL
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'TO3'),
        'Дизель генератор',
        'Проверка усилия нажатия на щетки'
    UNION ALL
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'TO3'),
        'Установка кабины',
        'Проверка испарителя на засорённость сот'
    UNION ALL
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'TO3'),
        'Установка кабины',
        'Замена фильтрующих элементов воздухозаборника кабины'
) t;

-- ============================================
-- 7. ДЕЙСТВИЯ ДЛЯ ПОСЛЕОБКАТОЧНОГО ТО
-- ============================================
INSERT INTO maintenance_actions (type_id, system_node, description, sort_order)
SELECT id, system_node, description, ROW_NUMBER() OVER (ORDER BY system_node)
FROM (
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'AFTER_RUN') as id,
        'Амортизатор' as system_node,
        'Проверить состояние резиновых амортизаторов дизель-генератора' as description
    UNION ALL
    SELECT 
        (SELECT id FROM maintenance_types WHERE code = 'AFTER_RUN'),
        'Установка дизель генератора',
        'Проверка вентилятора'
) t;

-- ============================================
-- 8. САМОСВАЛЫ
-- ============================================
INSERT INTO vehicles (vin, total_mileage, total_engine_hours, avg_speed)
VALUES
    ('Y3B7513DAT0000001', 87500.0, 6875.0, 12.73),
    ('Y3B7513DCT0000002', 102350.0, 7732.5, 13.24),
    ('Y3B7513DET0000003', 65750.0, 5304.0, 12.39),
    ('Y3B7513DHT0000004', 118420.0, 8756.3, 13.52),
    ('Y3B7513DJT0000005', 92180.0, 6844.4, 13.47),
    ('Y3B7513DKT0000006', 74300.0, 5834.6, 12.73),
    ('Y3B7513DLT0000007', 109620.0, 8282.5, 13.23),
    ('Y3B7513DPT0000008', 58900.0, 4773.1, 12.33),
    ('Y3B7513DTT0000009', 96850.0, 7334.8, 13.21),
    ('Y3B7513DVT0000010', 78450.0, 5940.0, 13.20);
