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

CREATE FUNCTION api.detail (p_base_item_name CITEXT, p_brand_name CITEXT) RETURNS JSONB AS $$
DECLARE
    p_result JSONB;
    p_variants JSONB;
    p_image_urls TEXT[];
    p_item_details JSONB;
BEGIN
    IF p_base_item_name IS NULL THEN
        RAISE EXCEPTION '"Base item name" is required';
    END IF;

    SELECT COALESCE(
        to_jsonb(bf),
        '{}'::jsonb
    )
    FROM (
        SELECT brand.name AS brand_name,
               base_item.name AS item_name,
               description,
               rating,
               image.url AS thumbnail_url
        FROM base_item
        JOIN brand USING (brand_id)
        LEFT JOIN image ON (base_item.thumbnail_image_id = image.image_id)
        WHERE base_item.name = p_base_item_name
      AND brand.brand_id = (SELECT brand_id FROM brand WHERE name = p_brand_name)

    ) bf INTO p_result;

    SELECT COALESCE(array_agg(url), '{}')
    FROM image JOIN base_item_image USING (image_id)
    WHERE base_item_image.base_item_id = (SELECT base_item_id FROM base_item WHERE name = p_base_item_name)
    INTO STRICT p_image_urls;

    p_result := p_result || jsonb_build_object('image_urls', p_image_urls);

    SELECT COALESCE(jsonb_agg(i), '[]'::jsonb)
    FROM item i
    WHERE i.base_item_id = (SELECT base_item_id FROM base_item WHERE name = p_base_item_name)
    INTO STRICT p_variants;

    p_result := p_result || jsonb_build_object('variants', p_variants);

    -- BEGIN KLUDGE
    -- This assumes it is always item.clothing
    SELECT COALESCE(jsonb_agg(to_jsonb(ic)), '[]'::jsonb)
    FROM item.clothing ic
    JOIN item i ON (ic.item_id = i.item_id)
    WHERE i.base_item_id = (SELECT base_item_id FROM base_item WHERE name = p_base_item_name)
    INTO STRICT p_item_details;

    RETURN p_result;
END;
$$ LANGUAGE plpgsql;
