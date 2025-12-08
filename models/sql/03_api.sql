CREATE SCHEMA api;

CREATE FUNCTION api.browse (
    p_page_index INTEGER,
    p_items_per_page INTEGER,
    p_include_tags TEXT[] DEFAULT NULL
) RETURNS JSONB AS $$
DECLARE
    v_items JSONB;
    v_total_count INTEGER;
    v_total_pages INTEGER;
    v_include_tag_ids INTEGER[];
BEGIN
    IF p_include_tags IS NOT NULL THEN
        v_include_tag_ids := ARRAY(
            SELECT tag_id FROM tag WHERE name = ANY(COALESCE(p_include_tags, '{}')::citext[])
        );
    END IF;

    v_items := (SELECT COALESCE(
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
        WHERE (
            v_include_tag_ids IS NULL
            OR (
            (SELECT COUNT(DISTINCT ti.tag_id)
             FROM tag_item ti
             WHERE ti.base_item_id = base_item.base_item_id
               AND ti.tag_id = ANY (v_include_tag_ids)
            ) =
            (SELECT COUNT(DISTINCT t) FROM unnest(v_include_tag_ids) t)
            )
        )
        LIMIT p_items_per_page OFFSET (p_page_index - 1) * p_items_per_page 
    ) bf);
    
    v_total_count := (SELECT COUNT(*)
            FROM base_item
        JOIN brand USING (brand_id)
        LEFT JOIN image ON (base_item.thumbnail_image_id = image.image_id)
        WHERE (
            v_include_tag_ids IS NULL
            OR (
            (SELECT COUNT(DISTINCT ti.tag_id)
             FROM tag_item ti
             WHERE ti.base_item_id = base_item.base_item_id
               AND ti.tag_id = ANY (v_include_tag_ids)
            ) =
            (SELECT COUNT(DISTINCT t) FROM unnest(v_include_tag_ids) t)
            )
        )
);
    v_total_pages := CEIL(v_total_count::DECIMAL / p_items_per_page);

    RETURN jsonb_build_object(
        'total_pages', v_total_pages,
        'items', v_items
    );
END;
$$ LANGUAGE plpgsql;

-- quick and dirty searching for tags, brands, and base items
CREATE FUNCTION api.search_bar (p_string CITEXT) RETURNS JSONB AS $$
DECLARE
    v_matching_tags TEXT[];
    v_matching_brands TEXT[];
    v_matching_items TEXT[];
    v_max_results INTEGER := 5;
BEGIN
    v_matching_tags := ARRAY(
        SELECT name FROM tag
        WHERE unaccent(name) ILIKE unaccent('%' || p_string || '%')
        ORDER BY
            CASE WHEN unaccent(name) ILIKE unaccent(p_string || '%') THEN 0 ELSE 1 END,
            unaccent(name)
        LIMIT v_max_results
    );
    v_matching_brands := ARRAY(
        SELECT name FROM brand
        WHERE unaccent(name) ILIKE unaccent('%' || p_string || '%')
        ORDER BY
            CASE WHEN unaccent(name) ILIKE unaccent(p_string || '%') THEN 0 ELSE 1 END,
            unaccent(name)
        LIMIT v_max_results
    );
    v_matching_items := ARRAY(
        SELECT name FROM base_item
        WHERE unaccent(name) ILIKE unaccent('%' || p_string || '%')
        ORDER BY
            CASE WHEN unaccent(name) ILIKE unaccent(p_string || '%') THEN 0 ELSE 1 END,
            unaccent(name)
        LIMIT v_max_results
    );

    RETURN jsonb_build_object(
        'tags', v_matching_tags,
        'brands', v_matching_brands,
        'items', v_matching_items
    );
END;
$$ LANGUAGE plpgsql;

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
                    'in_stock', COALESCE(inv.stock_quantity, 0) > 0
                ) ORDER BY bs.relative_order
            ),
            '[]'::jsonb
        )
        FROM item.clothing ic
        JOIN item i ON ic.item_id = i.item_id
        JOIN basic_size bs ON ic.basic_size = bs.size
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
CREATE FUNCTION api.transaction (
    p_transaction_event TEXT,
    p_item_id INTEGER,
    p_quantity INTEGER DEFAULT 1
) RETURNS VOID AS $$
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

-- Staff and admin are loaded through the command line
CREATE FUNCTION api.site_user_signup (
    p_first_name TEXT,
    p_last_name TEXT,
    p_username TEXT,
    p_email CITEXT,
    p_password TEXT
) RETURNS TEXT AS $$
BEGIN
    INSERT INTO site_user (first_name, last_name, username, email, password_hash, is_staff, is_admin)
    VALUES (
        p_first_name,
        p_last_name,
        p_username,
        p_email,
        crypt(p_password, gen_salt('bf')),
        FALSE,
        FALSE
    );

    RETURN api.site_user_authenticate(p_email, p_password);
END;
$$ LANGUAGE plpgsql;

CREATE FUNCTION api.site_user_authenticate (
    p_email CITEXT,
    p_password TEXT
) RETURNS TEXT AS $$
DECLARE
    v_password_hash TEXT;
    v_valid_password BOOLEAN;
    v_session_token TEXT;
BEGIN
    SELECT password_hash INTO v_password_hash
    FROM site_user
    WHERE email = p_email;
    IF v_password_hash IS NULL THEN
        RETURN NULL;
    END IF;
    
    v_valid_password := crypt(p_password, v_password_hash) = v_password_hash;
    IF NOT v_valid_password THEN
        RETURN NULL;
    END IF;

    -- remove any existing sessions for this user
    DELETE FROM session
    WHERE site_user_id = (SELECT site_user_id FROM site_user WHERE email = p_email);

    -- create new session
    INSERT INTO session (site_user_id, session_token, expires_at)
    VALUES (
        (SELECT site_user_id FROM site_user WHERE email = p_email),
        gen_random_uuid()::TEXT,
        NOW() + INTERVAL '1 day'
    ) RETURNING session_token INTO v_session_token;

    RETURN v_session_token;
END;
$$ LANGUAGE plpgsql;

CREATE FUNCTION api.user_validate_session (
    p_session_token TEXT
) RETURNS JSONB AS $$
DECLARE
    v_site_user_id INTEGER;
    v_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_count
    FROM session
    WHERE session_token = p_session_token
    AND expires_at > NOW();

    IF v_count <> 1 THEN
        RETURN NULL;
    END IF;

    RETURN jsonb_build_object(
        'first_name', su.first_name,
        'last_name', su.last_name,
        'username', su.username,
        'email', su.email,
        'is_staff', su.is_staff,
        'is_admin', su.is_admin,
        'created_at', su.created_at,
        'updated_at', su.updated_at
    )
    FROM site_user su
    JOIN session s ON su.site_user_id = s.site_user_id
    WHERE s.session_token = p_session_token;
END;
$$ LANGUAGE plpgsql;

CREATE FUNCTION api.user_signout (
    p_session_token TEXT
) RETURNS VOID AS $$
BEGIN
    DELETE FROM session
    WHERE session_token = p_session_token;
END;
$$ LANGUAGE plpgsql;