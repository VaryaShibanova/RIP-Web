-- Создание таблицы пользователей
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(25) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL,
    is_moderator BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы аномалий
CREATE TABLE anomalies (
    id SERIAL PRIMARY KEY,
    image VARCHAR(200),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    year INTEGER NOT NULL,
    pattern VARCHAR(200) NOT NULL,
    is_delete BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы деревьев (заявок)
CREATE TABLE trees (
    id SERIAL PRIMARY KEY,
    status VARCHAR(20) DEFAULT 'черновик' NOT NULL,
    description TEXT,
    total_rings INTEGER DEFAULT 0 NOT NULL,
    final_year INTEGER DEFAULT 0,
    date_create TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    date_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_finish TIMESTAMP NULL,
    creator_id INTEGER NOT NULL,
    moderator_id INTEGER NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (creator_id) REFERENCES users(id),
    FOREIGN KEY (moderator_id) REFERENCES users(id)
);

-- Создание таблицы элементов дерева
CREATE TABLE tree_items (
    id SERIAL PRIMARY KEY,
    tree_id INTEGER NOT NULL,
    anomaly_id INTEGER NOT NULL,
    anomalous_rings VARCHAR(50),
    calculated_year INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tree_id) REFERENCES trees(id) ON UPDATE CASCADE,
    FOREIGN KEY (anomaly_id) REFERENCES anomalies(id) ON UPDATE CASCADE,
    UNIQUE(tree_id, anomaly_id)
);

-- Вставка пользователей
INSERT INTO users (login, password, is_moderator) VALUES
('admin', '$2a$10$rOzJqT8bKQq3q3q3q3q3qO', true), -- password: admin123
('researcher', '$2a$10$rOzJqT8bKQq3q3q3q3q3qO', false), -- password: researcher123
('student', '$2a$10$rOzJqT8bKQq3q3q3q3q3qO', false); -- password: student123

-- Вставка аномалий (данные из первой лабораторной)
INSERT INTO anomalies (image, name, description, year, pattern, is_delete) VALUES
('http://127.0.0.1:9000/images/img/card1.jpg', 'Извержение вулкана Уайнапутина', 'На срезе дерева, жившего в 1600 году, хорошо заметно одно очень узкое и темное годовое кольцо. Оно резко контрастирует с более широкими светлыми кольцами до и после него. Это кольцо 1601 года.', 1600, 'Вулканическое извержение', false),
('http://127.0.0.1:9000/images/img/card2.jpg', 'Извержение Тамбора', 'На срезе дерева хорошо видно три аномальных кольца старше. Такие кольца 1815 года характеризуются узкими, фрагментированными темными кольцами 1816 года и относительно широкими кольцами 1817 года.', 1815, 'Вулканическое извержение', false),
('http://127.0.0.1:9000/images/img/card3.jpg', 'Извержение Каракатау', 'На срезе дерева, примерно на 15-16 кольцах от края, видны аномальные кольца второго следования 1884 года, например фрагментированное кольцо 1883 года и относительно широкими кольцами 1882 года.', 1883, 'Вулканическое извержение', false),
('http://127.0.0.1:9000/images/img/card4.jpg', 'Великое наводнение в Китае', 'Примерно на 50-90 кольцах от края видна аномальная необычная ширина, начинающаяся около 1931 года. Оно резко контрастирует с более узкими и темными кольцами предыдущих лет.', 1931, 'Наводнение', false),
('http://127.0.0.1:9000/images/img/card5.jpg', 'Наводнение Бурхарди', 'На срезе дерева, росшего в тот период, видно редкое очень широкое, с элементами черных волокон, кольцо соответствующее 1962 году или первому году после события. Оно заметно выделяется на фоне колец обычной ширины.', 1962, 'Наводнение', false),
('http://127.0.0.1:9000/images/img/card6.jpg', 'Год без лета', 'Чёткая тёмная полоса 1816 года — «год без лета» — выглядит как шрам, врезавшийся в память дерева. Это сверхузкое, почти чёрное кольцо. Рядом видны более светлые и широкие кольца.', 1816, 'Климатическая аномалия', false),
('http://127.0.0.1:9000/images/img/card7.jpg', 'Великий пожар годов Мэйрэки', 'На срезе вида аномалия на 15-м кольце от коры. Само кольцо 1657 года неровное и фрагментированное: с одной стороны узкое и плотное, с другой — более широкое. За ним следует аномально широкое светлое кольцо 1658 года.', 1657, 'Лесной пожар', false);

-- Вставка деревьев (заявок) с разными статусами
INSERT INTO trees (status, description, total_rings, final_year, date_create, date_update, creator_id) VALUES
('черновик', 'Исследование вулканических событий по древесным кольцам', 235, 1600, '2024-01-15 10:30:00', '2024-01-15 10:30:00', 2),
('сформирован', 'Анализ климатических изменений в XVII веке', 180, 1657, '2024-01-10 14:20:00', '2024-01-12 16:45:00', 2),
('завершён', 'Комплексное исследование аномалий XIX века', 220, 1883, '2024-01-05 09:15:00', '2024-01-08 11:30:00', 3),
('отклонён', 'Исследование наводнений XX века', 150, 1931, '2024-01-03 13:40:00', '2024-01-04 15:20:00', 3),
('удалён', 'Предварительное исследование', 100, 1962, '2023-12-20 08:50:00', '2023-12-21 10:30:00', 2);

-- Вставка элементов деревьев
INSERT INTO tree_items (tree_id, anomaly_id, anomalous_rings, calculated_year) VALUES
(1, 1, '3,4,5', 1600),
(1, 2, '15,16,17', 1815),
(2, 7, '12-15', 1657),
(3, 3, '25-30', 1883),
(3, 6, '18-20', 1816),
(4, 4, '45-50', 1931),
(5, 5, '60-65', 1962);

-- Создание индексов для улучшения производительности
CREATE INDEX idx_anomalies_name ON anomalies(name);
CREATE INDEX idx_anomalies_year ON anomalies(year);
CREATE INDEX idx_trees_status ON trees(status);
CREATE INDEX idx_trees_creator_id ON trees(creator_id);
CREATE INDEX idx_tree_items_tree_id ON tree_items(tree_id);
CREATE INDEX idx_tree_items_anomaly_id ON tree_items(anomaly_id);

-- Создание представления для удобного просмотра деревьев с аномалиями
CREATE VIEW tree_details AS
SELECT 
    t.id as tree_id,
    t.status,
    t.description as tree_description,
    t.total_rings,
    t.final_year,
    t.date_create,
    u.login as creator_login,
    ti.anomalous_rings,
    ti.calculated_year,
    a.name as anomaly_name,
    a.description as anomaly_description,
    a.year as anomaly_year,
    a.image as anomaly_image
FROM trees t
LEFT JOIN users u ON t.creator_id = u.id
LEFT JOIN tree_items ti ON t.id = ti.tree_id
LEFT JOIN anomalies a ON ti.anomaly_id = a.id
WHERE t.status != 'удалён';

-- Комментарии к таблицам
COMMENT ON TABLE users IS 'Таблица пользователей системы';
COMMENT ON TABLE anomalies IS 'Таблица аномальных паттернов в древесных кольцах';
COMMENT ON TABLE trees IS 'Таблица заявок на исследование (деревьев)';
COMMENT ON TABLE tree_items IS 'Связующая таблица деревьев и аномалий';

COMMENT ON COLUMN trees.status IS 'Статус заявки: черновик, сформирован, завершён, отклонён, удалён';
COMMENT ON COLUMN trees.total_rings IS 'Общее количество колец в исследовании';
COMMENT ON COLUMN trees.final_year IS 'Итоговый год исследования';
COMMENT ON COLUMN tree_items.anomalous_rings IS 'Номера аномальных колец (например: 3,4,5 или 15-20)';
COMMENT ON COLUMN tree_items.calculated_year IS 'Рассчитанный год аномалии';