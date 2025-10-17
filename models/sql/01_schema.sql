SET
    client_encoding = 'UTF8';

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE basic_size (size CITEXT PRIMARY KEY);

INSERT INTO
    basic_size (size)
VALUES
    ('XS'),
    ('S'),
    ('M'),
    ('L'),
    ('XL'),
    ('XXL'),
    ('XXXL');

CREATE TABLE color (
    color_id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    hex_code TEXT NOT NULL
);

INSERT INTO
    color (name, hex_code)
VALUES
    ('Gold', '#D4AF37'),
    ('Silver', '#C2C2C2'),
    ('White', '#FFFBF0'),
    ('Black', '#0B0B0B'),
    ('Red', '#8B0000'),
    ('Blue', '#0F52BA'),
    ('Green', '#046307'),
    ('Yellow', '#F7E7CE'),
    ('Purple', '#6A0DAD'),
    ('Pink', '#F2C1D1'),
    ('Brown', '#5C2F1F'),
    ('Gray', '#6E6E6E');

CREATE TABLE brand (
    brand_id SERIAL PRIMARY KEY,
    name CITEXT UNIQUE NOT NULL
);

CREATE TABLE tag (
    tag_id SERIAL PRIMARY KEY,
    parent_tag_id INTEGER REFERENCES tag (tag_id) ON DELETE CASCADE,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE image (
    image_id SERIAL PRIMARY KEY,
    url TEXT UNIQUE NOT NULL,
    added TIMESTAMPTZ DEFAULT NOW(),
    alt TEXT
);

CREATE TABLE base_item (
    base_item_id SERIAL PRIMARY KEY,
    name CITEXT NOT NULL,
    description TEXT,
    brand_id INTEGER REFERENCES brand (brand_id),
    thumbnail_image_id INTEGER REFERENCES image (image_id),
    rating NUMERIC(2, 1) CHECK (
        rating >= 0
        AND rating <= 5
    ),
    added TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE base_item_image (
    base_item_id INTEGER REFERENCES base_item (base_item_id) ON DELETE CASCADE,
    image_id INTEGER REFERENCES image (image_id) ON DELETE CASCADE,
    PRIMARY KEY (base_item_id, image_id)
);

CREATE FUNCTION add_base_item (
    p_name TEXT,
    p_description TEXT,
    p_brand_name TEXT,
    p_thumbnail_url TEXT,
    p_image_urls TEXT[] DEFAULT '{}'::text[],
    p_rating NUMERIC(2, 1) DEFAULT NULL
) RETURNS INT AS $$
DECLARE
    v_brand_id INTEGER;
    v_base_item_id INTEGER;
    v_image_id INTEGER;
    v_url TEXT;
BEGIN
    IF p_brand_name IS NOT NULL THEN
        v_brand_id := (SELECT brand_id FROM brand WHERE name = p_brand_name);
    END IF;

    IF p_thumbnail_url IS NOT NULL THEN
        INSERT INTO image (url) VALUES (p_thumbnail_url) ON CONFLICT (url) DO NOTHING RETURNING image_id INTO v_image_id;
    END IF;

    INSERT INTO base_item (name, description, brand_id, thumbnail_image_id, rating) VALUES (p_name, p_description, v_brand_id, v_image_id, p_rating) RETURNING base_item_id INTO v_base_item_id;

    FOREACH v_url IN ARRAY p_image_urls
    LOOP
        INSERT INTO image (url) VALUES (v_url) ON CONFLICT (url) DO NOTHING RETURNING image_id INTO v_image_id;
        IF v_image_id IS NULL THEN
            SELECT image_id FROM image WHERE url = v_url INTO v_image_id;
        END IF;
        INSERT INTO base_item_image (base_item_id, image_id) VALUES (v_base_item_id, v_image_id);
    END LOOP;

    RETURN v_base_item_id;
END;
$$ LANGUAGE plpgsql;

-- This is something that physically exists in inventory.  Should not be used directly
CREATE TABLE item (
    item_id SERIAL PRIMARY KEY,
    base_item_id INTEGER REFERENCES base_item (base_item_id) ON DELETE CASCADE
);

CREATE TABLE transaction_event (transaction_event TEXT PRIMARY KEY);

INSERT INTO
    transaction_event (transaction_event)
VALUES
    -- sets the inventory to the audited quantity
    ('audit'),
    -- adds inventory
    ('ship-in'),
    -- removes inventory
    ('ship-out'),
    -- removes inventory
    ('scrap');

CREATE TABLE inventory_transaction (
    inventory_transaction_id serial PRIMARY KEY,
    item_id INT NOT NULL REFERENCES item (item_id) ON DELETE RESTRICT,
    -- for nested transactions, like building an assembly that consumes parts...
    -- parent_transaction_id INT REFERENCES inventory_transaction (inventory_transaction_id) ON DELETE CASCADE,
    transaction_event TEXT NOT NULL REFERENCES transaction_event (transaction_event) ON DELETE RESTRICT,
    transaction_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    delta_quantity INT NOT NULL
);

CREATE VIEW inventory AS
SELECT
    it.item_id,
    SUM(it.delta_quantity) AS stock_quantity
FROM
    inventory_transaction it
GROUP BY
    it.item_id;
-- HAVING
--     SUM(it.delta_quantity) > 0;


-- Specific types of items go in their own schema
CREATE SCHEMA item;

CREATE TABLE item.clothing (
    item_id INTEGER PRIMARY KEY REFERENCES item (item_id) ON DELETE CASCADE,
    basic_size CITEXT REFERENCES basic_size (size),
    -- color_id INTEGER REFERENCES color (color_id),
    -- UNIQUE (item_id, basic_size, color_id)
    UNIQUE (item_id, basic_size)
);

CREATE FUNCTION add_clothing_item (
    p_base_item_id INTEGER,
    p_basic_size CITEXT
) RETURNS INT AS $$
DECLARE
    v_item_id INTEGER;
BEGIN
    INSERT INTO item (base_item_id) VALUES (p_base_item_id) RETURNING item_id INTO v_item_id;
    INSERT INTO item.clothing (item_id, basic_size) VALUES (v_item_id, p_basic_size);
    RETURN v_item_id;
END;
$$ LANGUAGE plpgsql;

-- CREATE TABLE item.shoes (
--     shoes_id SERIAL PRIMARY KEY,
--     item_id INTEGER REFERENCES item (item_id) ON DELETE CASCADE,
--     size NUMERIC(3, 1) NOT NULL,
--     UNIQUE (item_id, size)
-- );

-- CREATE TABLE item.bottom (
--     bottom_id SERIAL PRIMARY KEY,
--     item_id INTEGER REFERENCES item (item_id) ON DELETE CASCADE,
--     waist INTEGER NOT NULL,
--     inseam INTEGER NOT NULL,
--     color_id INTEGER REFERENCES color (color_id),
--     UNIQUE (item_id, waist, inseam, color_id)
-- );
