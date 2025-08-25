set check_function_bodies = off;

CREATE OR REPLACE FUNCTION public.create_client_and_secret(p_display text, p_env_id uuid)
 RETURNS TABLE(client_id uuid, secret_id uuid, secret text)
 SECURITY DEFINER
 LANGUAGE plpgsql
 SET search_path TO 'public', 'extensions'
AS $function$BEGIN

    -- must be admin
    IF NOT (is_admin_client()) THEN
        RAISE EXCEPTION 'unauthorized';
    END IF;

    -- create client
    INSERT INTO public.clients (display, environment_id)
    VALUES (p_display, p_env_id)
    RETURNING id INTO client_id;

    -- create the secret with a random password
    secret := encode(extensions.gen_random_bytes(24), 'base64'); -- ~32 chars, base64-safe
    secret_id := public.create_client_secret(client_id, secret);
    RETURN NEXT;
END;$function$
;

CREATE OR REPLACE FUNCTION public.create_client_secret(p_client_id uuid, p_secret text)
 RETURNS uuid
 LANGUAGE plpgsql
 SECURITY DEFINER
 SET search_path TO 'public', 'extensions'
AS $function$DECLARE
    secret_id uuid;
BEGIN

    -- must be admin
    IF NOT (is_admin_client()) THEN
        RAISE EXCEPTION 'unauthorized';
    END IF;

    INSERT INTO public.clients_secrets (hash, client_id)
    VALUES (extensions.crypt(p_secret, extensions.gen_salt('bf', 12)), -- bcrypt cost 12
            p_client_id)
    RETURNING id INTO secret_id;
    RETURN secret_id;
END;$function$
;


