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
               thumbnail_url
        FROM base_item
        JOIN brand USING (brand_id)
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
