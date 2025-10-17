CREATE SCHEMA api;

CREATE FUNCTION api.browse (
    p_page_index INTEGER DEFAULT 1,
    p_items_per_page INTEGER DEFAULT 20
) RETURNS JSONB AS $$
DECLARE
    items JSONB;
    total_count INTEGER;
    current_page INTEGER;
    total_pages INTEGER;
BEGIN
    SELECT COALESCE(
        jsonb_agg(to_jsonb(bf)),
        '[]'::jsonb
    )
    FROM (
        SELECT brand.name AS brand_name,
               base_item.name AS item_name,
               description,
               image.url AS thumbnail_url
        FROM base_item
        JOIN brand USING (brand_id)
        LEFT JOIN image ON (base_item.thumbnail_image_id = image.image_id)
        LIMIT p_items_per_page OFFSET (p_page_index - 1) * p_items_per_page 
    ) bf INTO items;
    
    SELECT COUNT(*) FROM base_item INTO total_count;
    total_pages := CEIL(total_count::DECIMAL / p_items_per_page);

    RETURN jsonb_build_object(
        'total_pages', total_pages,
        'items', items
    );
END;
$$ LANGUAGE plpgsql;

CREATE EXTENSION IF NOT EXISTS unaccent;

CREATE FUNCTION api.brands () RETURNS JSONB AS $$
DECLARE
    brands JSONB;
BEGIN
    SELECT COALESCE(
        jsonb_object_agg(
            initial,
            brand_names
        ),
        '{}'::jsonb
    )
    FROM (
        SELECT CASE
                 WHEN SUBSTRING(unaccent(name), 1, 1) ~ '^[0-9]' THEN '#'
                 ELSE UPPER(SUBSTRING(unaccent(name), 1, 1))
               END AS initial,
               jsonb_agg(name ORDER BY unaccent(name)) AS brand_names
        FROM brand
        GROUP BY initial
        ORDER BY initial
    ) b INTO brands;
    RETURN brands;
END;
$$ LANGUAGE plpgsql;

CREATE FUNCTION api.detail (
    p_base_item_name CITEXT,
    p_brand_name CITEXT DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_base_item_id INTEGER;
    v_brand_name CITEXT;
    v_description TEXT;
    v_rating NUMERIC(2, 1);
    v_image_urls TEXT[];
    v_item_specific_details JSONB;
BEGIN
    IF p_base_item_name IS NULL THEN
        RAISE EXCEPTION '"Base item name" is required';
    END IF;

    -- if brand is not provided
    v_brand_name := p_brand_name;
    IF v_brand_name IS NULL THEN
        SELECT brand.name FROM brand
        JOIN base_item USING (brand_id)
        WHERE base_item.name = p_base_item_name
        LIMIT 1 INTO v_brand_name;
    END IF;
    IF v_brand_name IS NULL THEN
        RAISE EXCEPTION 'Could not determine brand for base item "%" - please provide brand name', p_base_item_name;
    END IF;
    
    v_base_item_id := (SELECT base_item_id FROM base_item
                       JOIN brand USING (brand_id)
                       WHERE base_item.name = p_base_item_name
                         AND brand.name = v_brand_name);
    IF v_base_item_id IS NULL THEN
        RAISE EXCEPTION 'Could not find base item "%" for brand "%"', p_base_item_name, v_brand_name;
    END IF;

    v_description := (SELECT description FROM base_item WHERE base_item_id = v_base_item_id);   

    v_image_urls := (SELECT COALESCE(array_agg(url), '{}')
    FROM image JOIN base_item_image USING (image_id)
    WHERE base_item_image.base_item_id = (SELECT base_item_id FROM base_item WHERE name = p_base_item_name));

    v_rating := (SELECT rating FROM base_item WHERE base_item_id = v_base_item_id);

    -- BEGIN KLUDGE
    -- This assumes it is always item.clothing
    v_item_specific_details := (
        SELECT COALESCE(
            jsonb_agg(
                jsonb_build_object(
                    'size', ic.basic_size,
                    'stock_quantity', COALESCE(inv.stock_quantity, 0)
                )
            ),
            '[]'::jsonb
        )
        FROM item.clothing ic
        JOIN item i ON ic.item_id = i.item_id
        LEFT JOIN inventory inv ON i.item_id = inv.item_id
        WHERE i.base_item_id = v_base_item_id
    ); -- END KLUDGE

    RETURN jsonb_build_object(
        'item_name', p_base_item_name,
        'brand_name', v_brand_name,
        'description', v_description,
        'image_urls', v_image_urls,
        'rating', v_rating,
        'item_specific_details', v_item_specific_details
    );
END;
$$ LANGUAGE plpgsql;

-- Modify inventory
CREATE FUNCTION api.transaction(p_transaction_event TEXT, p_item_id INTEGER, p_quantity INTEGER DEFAULT 1) 
RETURNS VOID AS $$
DECLARE
    v_delta_quantity INTEGER;
    v_current_quantity INTEGER;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM transaction_event WHERE transaction_event = p_transaction_event)
    THEN
        RAISE EXCEPTION 'Invalid transaction event: %', p_transaction_event;
    END IF;

    IF p_transaction_event = 'audit' THEN
        IF p_quantity < 0 THEN
            RAISE EXCEPTION 'Audit sets inventory to this value.  It cannot be negative: %', p_quantity;
        END IF;
        v_current_quantity := (SELECT stock_quantity FROM inventory WHERE item_id = p_item_id);
        v_delta_quantity := p_quantity - COALESCE(v_current_quantity, 0);
    
    ELSIF p_transaction_event = 'ship-out' OR p_transaction_event = 'scrap' THEN
        v_delta_quantity := -p_quantity;
    
    ELSE
        v_delta_quantity := p_quantity;
    END IF;

    INSERT INTO inventory_transaction (transaction_event, item_id, delta_quantity)
    VALUES (p_transaction_event, p_item_id, v_delta_quantity);
END;
$$ LANGUAGE plpgsql;