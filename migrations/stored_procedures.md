## createprofile
```

CREATE OR REPLACE PROCEDURE public.createprofile(
	IN p_oid uuid,
	IN p_nickname character varying,
	IN p_first_name character varying,
	IN p_last_name character varying,
	IN p_password character varying,
	IN p_created_at timestamp with time zone,
	IN p_updated_at timestamp with time zone,
	IN p_state integer,
	IN p_user_role integer)
LANGUAGE 'plpgsql'
AS $BODY$
INSERT INTO user_profiles (oid, nickname, first_name, last_name, password, created_at, updated_at, state, user_role)
VALUES (p_oid, p_nickname, p_first_name, p_last_name, p_password, p_created_at, p_updated_at, p_state, p_user_role);
$BODY$;
ALTER PROCEDURE public.createprofile(uuid, character varying, character varying, character varying, character varying, timestamp with time zone, timestamp with time zone, integer, integer)
    OWNER TO postgres;

```
## delete_user
```

CREATE OR REPLACE PROCEDURE public.delete_user(
	IN p_oid uuid)
LANGUAGE 'plpgsql'
AS $BODY$
BEGIN
    DELETE FROM user_profiles
    WHERE oid = p_oid;
END;
$BODY$;
ALTER PROCEDURE public.delete_user(uuid)
    OWNER TO postgres;

```

## get_count
```

CREATE OR REPLACE PROCEDURE public.get_count(
	OUT result integer)
LANGUAGE 'plpgsql'
AS $BODY$
BEGIN
    SELECT COUNT(*) INTO result
    FROM user_profiles;
END;
$BODY$;
ALTER PROCEDURE public.get_count()
    OWNER TO postgres;

```

## get_user
```

CREATE OR REPLACE PROCEDURE public.get_user(
	IN p_oid uuid,
	OUT p_nickname character varying,
	OUT p_first_name character varying,
	OUT p_last_name character varying,
	OUT p_created_at timestamp without time zone,
	OUT p_updated_at timestamp without time zone,
	OUT p_state character varying)
LANGUAGE 'plpgsql'
AS $BODY$
BEGIN
    SELECT nickname, first_name, last_name, created_at, updated_at, state
    INTO p_nickname, p_first_name, p_last_name, p_created_at, p_updated_at, p_state
    FROM user_profiles
    WHERE oid = p_oid;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'User profile with oid % not found', p_oid;
    END IF;
END;
$BODY$;
ALTER PROCEDURE public.get_user(uuid)
    OWNER TO postgres

```

## FUNCTION get_all_users
```

CREATE OR REPLACE FUNCTION public.get_all_users(p_limit INT, p_offset INT)
RETURNS TABLE (
    p_oid UUID,
    p_nickname VARCHAR(255),
    p_first_name VARCHAR(255),
    p_last_name VARCHAR(255),
    p_created_at TIMESTAMP,
    p_updated_at TIMESTAMP,
    p_state INTEGER)
AS $$
BEGIN
    RETURN QUERY
    SELECT oid, nickname, first_name, last_name, created_at::TIMESTAMP, updated_at::TIMESTAMP, state
    FROM user_profiles
    ORDER BY created_at
    LIMIT p_limit
    OFFSET p_offset;
END;
$$ LANGUAGE plpgsql;

```

## get_user_for_token
```

CREATE OR REPLACE PROCEDURE public.get_user_for_token(
	IN p_nickname character varying,
	OUT p_oid uuid,
	OUT p_nickname_out character varying,
	OUT p_user_role character varying,
	OUT p_state character varying)
LANGUAGE 'plpgsql'
AS $BODY$
BEGIN
    SELECT oid, nickname, user_role, state
    INTO p_oid, p_nickname_out, p_user_role, p_state
    FROM user_profiles
    WHERE nickname = p_nickname;
END;
$BODY$;
ALTER PROCEDURE public.get_user_for_token(character varying)
    OWNER TO postgres;

```

## get_user_password
```

CREATE OR REPLACE PROCEDURE public.get_user_password(
	IN p_nickname character varying,
	OUT p_password_hash character varying)
LANGUAGE 'plpgsql'
AS $BODY$
BEGIN
    SELECT password
    INTO p_password_hash
    FROM user_profiles
    WHERE nickname = p_nickname;
END;
$BODY$;
ALTER PROCEDURE public.get_user_password(character varying)
    OWNER TO postgres;

```

## get_user_state
```

CREATE OR REPLACE PROCEDURE public.get_user_state(
	IN p_oid uuid,
	OUT p_state character varying)
LANGUAGE 'plpgsql'
AS $BODY$
BEGIN
    SELECT state
    INTO p_state
    FROM user_profiles
    WHERE oid = p_oid;
END;
$BODY$;
ALTER PROCEDURE public.get_user_state(uuid)
    OWNER TO postgres;

```

## getuser
```

CREATE OR REPLACE PROCEDURE public.getuser(
	IN p_oid uuid,
	OUT p_nickname character varying,
	OUT p_first_name character varying,
	OUT p_last_name character varying,
	OUT p_created_at timestamp without time zone,
	OUT p_updated_at timestamp without time zone,
	OUT p_state character varying)
LANGUAGE 'plpgsql'
AS $BODY$
BEGIN
    SELECT nickname, first_name, last_name, created_at, updated_at, state
    INTO p_nickname, p_first_name, p_last_name, p_created_at, p_updated_at, p_state
    FROM user_profiles
    WHERE oid = p_oid;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'User profile with oid % not found', p_oid;
    END IF;
END;
$BODY$;
ALTER PROCEDURE public.getuser(uuid)
    OWNER TO postgres;

```

## update_password
```

CREATE OR REPLACE PROCEDURE public.update_password(
	IN p_password character varying,
	IN p_updated_at timestamp with time zone,
	IN p_oid uuid)
LANGUAGE 'sql'
AS $BODY$
UPDATE user_profiles
SET password=p_password, updated_at=p_updated_at
WHERE oid=p_oid;
$BODY$;
ALTER PROCEDURE public.update_password(character varying, timestamp with time zone, uuid)
    OWNER TO postgres;

```

## update_profile
```

CREATE OR REPLACE PROCEDURE public.update_profile(
	IN p_nickname character varying,
	IN p_first_name character varying,
	IN p_last_name character varying,
	IN p_updated_at timestamp with time zone,
	IN p_oid uuid)
LANGUAGE 'sql'
AS $BODY$
UPDATE user_profiles
SET nickname=p_nickname, first_name=p_first_name, last_name=p_last_name, updated_at=p_updated_at
WHERE oid=p_oid;
$BODY$;
ALTER PROCEDURE public.update_profile(character varying, character varying, character varying, timestamp with time zone, uuid)
    OWNER TO postgres;

```