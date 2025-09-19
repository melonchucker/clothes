CREATE FUNCTION load_tags (p_tags JSONB, p_parent_name TEXT DEFAULT NULL) RETURNS VOID AS $$
DECLARE
    v_tag_name TEXT;
    v_subtags JSONB;
BEGIN
    FOR v_tag_name, v_subtags IN SELECT * FROM jsonb_each(p_tags)
    LOOP
        PERFORM add_tag (v_tag_name, p_parent_name);

        IF jsonb_typeof(v_subtags) = 'object' THEN
            PERFORM load_tags (v_subtags, v_tag_name);
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

CREATE FUNCTION load_dummy_data () RETURNS VOID AS $$
DECLARE
    v_tags JSONB;
    v_brands TEXT[];
    v_brand_name TEXT;
    v_brand_id INTEGER;
BEGIN
    v_tags := '
        {
            "Accessories": {
                "Hats": {},
                "Bags": {},
                "Jewelry": {}
            },
            "Clothing": {
                "Bottoms": {
                    "Jeans": {},
                    "Pants": {},
                    "Shorts": {},
                    "Skirts": {}
                },
                "Tops": {
                    "Blouses": {},
                    "Shirts": {},
                    "Sweaters": {},
                    "Jumpsuits & Rompers": {},
                    "Tank Tops": {}
                },
                "Dresses": {},
                "Outerwear": {
                    "Coats": {},
                    "Jackets": {},
                    "Blazers": {},
                    "Vests": {},
                    "Cardigans": {}
                },
                "Intimates": {},
                "Activewear": {}
            },
            "Shoes": {
                "Flats": {},
                "Heels": {},
                "Boots": {},
                "Sandals": {},
                "Sneakers": {}
            },
            "Seasonal": {
                "Spring": {
                    "Easter": {}
                },
                "Summer": {
                    "Fourth of July": {}
                },
                "Fall": {
                    "Halloween": {},
                    "Thanksgiving": {}
                },
                "Winter": {
                    "Christmas": {},
                    "New Years": {},
                    "Valentines Day": {}
                }
            },
            "Style": {
                "Casual": {},
                "Formal": {},
                "Business Casual": {},
                "Boho": {},
                "Preppy": {},
                "Streetwear": {},
                "Athleisure": {},
                "Vintage": {}
            }
        }';

    PERFORM load_tags (v_tags);
END;
$$ LANGUAGE plpgsql;

SELECT
    load_dummy_data ();