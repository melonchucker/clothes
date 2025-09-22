CREATE SCHEMA api;

CREATE FUNCTION api.browse(p_limit INTEGER DEFAULT 20, p_offset INTEGER DEFAULT 0) RETURNS JSONB AS $$
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
        LIMIT p_limit OFFSET p_offset
    ) bf;
$$ LANGUAGE sql;
