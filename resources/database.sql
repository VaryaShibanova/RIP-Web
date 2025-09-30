-- Настройка базы данных для лабораторной работы 3
-- Дендрохронологическое исследование

-- Очистка существующих данных (опционально, если нужно пересоздать)
-- DELETE FROM tree_items;
-- DELETE FROM trees;
-- DELETE FROM anomalies;
-- DELETE FROM users;

-- Пользователи
INSERT INTO users (id, login, password, is_moderator) VALUES
(1, 'research_user', 'password123', false),
(2, 'moderator_user', 'password123', true),
(3, 'scientist_user', 'password123', false)
ON CONFLICT (id) DO UPDATE SET
    login = EXCLUDED.login,
    password = EXCLUDED.password,
    is_moderator = EXCLUDED.is_moderator;

-- Аномалии (услуги) - реальные данные для дендрохронологии
INSERT INTO anomalies (id, is_delete, image, name, description, year) VALUES
(1, false, 'http://127.0.0.1:9000/images/img/card1.jpg', 'Извержение вулкана Уайнапутина', 'На срезе дерева, жившего в 1600 году, хорошо заметно одно очень узкое и темное годовое кольцо. Оно резко контрастирует с более широкими светлыми кольцами до и после него. Это кольцо 1601 года.', 1600),
(2, false, 'http://127.0.0.1:9000/images/img/card2.jpg', 'Извержение Тамбора', 'На срезе дерева хорошо видно три аномальных кольца старше. Такие кольца 1815 года характеризуются узкими, фрагментированными темными кольцами 1816 года и относительно широкими кольцами 1817 года.', 1815),
(3, false, 'http://127.0.0.1:9000/images/img/card3.jpg', 'Извержение Кракатау', 'На срезе дерева, примерно на 15-16 кольцах от края, видны аномальные кольца второго следования 1884 года, например фрагментированное кольцо 1883 года и относительно широкими кольцами 1882 года.', 1883),
(4, false, 'http://127.0.0.1:9000/images/img/card4.jpg', 'Великое наводнение в Китае', 'Примерно на 50-90 кольцах от края видна аномальная необычная ширина, начинающаяся около 1931 года. Оно резко контрастирует с более узкими и темными кольцами предыдущих лет.', 1931),
(5, false, 'http://127.0.0.1:9000/images/img/card5.jpg', 'Наводнение Бурхарди', 'На срезе дерева, росшего в тот период, видно редкое очень широкое, с элементами черных волокон, кольцо соответствующее 1962 году или первому году после события. Оно заметно выделяется на фоне колец обычной ширины.', 1962),
(6, false, 'http://127.0.0.1:9000/images/img/card6.jpg', 'Год без лета', 'Чёткая тёмная полоса 1816 года — «год без лета» — выглядит как шрам, врезавшийся в память дерева. Это сверхузкое, почти чёрное кольцо. Рядом видны более светлые и широкие кольца.', 1816),
(7, false, 'http://127.0.0.1:9000/images/img/card7.jpg', 'Великий пожар годов Мэйрэки', 'На срезе вида аномалия на 15-м кольце от коры. Само кольцо 1657 года неровное и фрагментированное: с одной стороны узкое и плотное, с другой — более широкое. За ним следует аномально широкое светлое кольцо 1658 года.', 1657)
ON CONFLICT (id) DO UPDATE SET
    is_delete = EXCLUDED.is_delete,
    image = EXCLUDED.image,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    year = EXCLUDED.year;

-- Заявки (деревья) для исследований
INSERT INTO trees (id, status, description, total_rings, final_year, date_create, date_update, date_finish, creator_id, moderator_id) VALUES
(1, 'черновик', 'Исследование вулканических аномалий в Сибири', 150, 0, '2024-01-10 10:00:00', '2024-01-10 10:00:00', NULL, 1, NULL),
(2, 'сформирован', 'Анализ климатических изменений по годичным кольцам', 200, 0, '2024-01-08 14:30:00', '2024-01-09 09:15:00', NULL, 1, NULL),
(3, 'завершён', 'Комплексное исследование исторических событий по дендрохронологии', 180, 1642, '2024-01-05 08:00:00', '2024-01-07 16:45:00', '2024-01-07 16:45:00', 3, 2),
(4, 'отклонён', 'Изучение локальных погодных аномалий', 120, 0, '2024-01-03 11:20:00', '2024-01-04 13:10:00', '2024-01-04 13:10:00', 1, 2),
(5, 'удалён', 'Архивное исследование древних образцов', 90, 0, '2024-01-01 09:00:00', '2024-01-02 12:00:00', NULL, 3, NULL),
(6, 'черновик', 'Новое исследование вулканического воздействия', 0, 0, '2024-01-15 08:30:00', '2024-01-15 08:30:00', NULL, 1, NULL)
ON CONFLICT (id) DO UPDATE SET
    status = EXCLUDED.status,
    description = EXCLUDED.description,
    total_rings = EXCLUDED.total_rings,
    final_year = EXCLUDED.final_year,
    date_create = EXCLUDED.date_create,
    date_update = EXCLUDED.date_update,
    date_finish = EXCLUDED.date_finish,
    creator_id = EXCLUDED.creator_id,
    moderator_id = EXCLUDED.moderator_id;

-- Элементы заявок (м-м связь) с реальными данными по кольцам
INSERT INTO tree_items (id, tree_id, anomaly_id, anomalous_rings, calculated_year) VALUES
(1, 1, 1, '45,67,89', 1638),
(2, 1, 2, '112,125', 1845),
(3, 2, 3, '78,92,105', 1899),
(4, 2, 4, '34,56', 1955),
(5, 3, 1, '23,45,67,89,101', 1642),
(6, 3, 2, '112,134', 1849),
(7, 3, 4, '45,67', 1959),
(8, 4, 5, '15,28,42', 1978),
(9, 3, 6, '88,99,110', 1832),
(10, 2, 7, '25,36,47', 1673)
ON CONFLICT (id) DO UPDATE SET
    tree_id = EXCLUDED.tree_id,
    anomaly_id = EXCLUDED.anomaly_id,
    anomalous_rings = EXCLUDED.anomalous_rings,
    calculated_year = EXCLUDED.calculated_year;

-- Сброс последовательностей
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users), true);
SELECT setval('anomalies_id_seq', (SELECT MAX(id) FROM anomalies), true);
SELECT setval('trees_id_seq', (SELECT MAX(id) FROM trees), true);
SELECT setval('tree_items_id_seq', (SELECT MAX(id) FROM tree_items), true);

-- Проверка данных
SELECT 'Users:' as "Проверка данных";
SELECT id, login, is_moderator FROM users;

SELECT 'Anomalies:' as "";
SELECT id, name, year FROM anomalies WHERE is_delete = false;

SELECT 'Trees:' as "";
SELECT id, status, description, total_rings FROM trees WHERE status != 'удалён';

SELECT 'Tree Items:' as "";
SELECT ti.id, t.id as tree_id, a.name as anomaly_name, ti.anomalous_rings, ti.calculated_year 
FROM tree_items ti
JOIN trees t ON ti.tree_id = t.id
JOIN anomalies a ON ti.anomaly_id = a.id;