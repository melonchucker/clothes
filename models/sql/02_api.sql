CREATE SCHEMA api;

CREATE FUNCTION api.get_base_items() RETURNS JSONB AS $$
BEGIN
    RETURN (
        SELECT COALESCE(
            jsonb_agg(
                to_jsonb(bi) || jsonb_build_object(
                    'brand', b.name,
                    'pictures', COALESCE(pics.pictures, '[]'::jsonb)
                )
            ),
            '[]'::jsonb
        )
        FROM base_item bi
        JOIN brand b ON bi.brand_id = b.brand_id
        LEFT JOIN (
            SELECT base_item_id, jsonb_agg(to_jsonb(bip) - 'base_item_id') AS pictures
            FROM base_item_picture bip
            GROUP BY base_item_id
        ) pics ON pics.base_item_id = bi.base_item_id
    );
END;
$$ LANGUAGE plpgsql;