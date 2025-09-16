set check_function_bodies = off;

CREATE OR REPLACE FUNCTION public.secrets()
 RETURNS TABLE(id uuid, encrypted_secret text, decrypted_secret text, created_at timestamp with time zone, updated_at timestamp with time zone)
 LANGUAGE plpgsql
 STABLE SECURITY DEFINER
 SET search_path TO ''
AS $function$
BEGIN

  -- must be admin
  IF NOT (private.is_admin_client()) THEN
    RAISE EXCEPTION 'unauthorized';
  END IF;

  RETURN QUERY
  SELECT
    s.id,
    s.secret AS encrypted_secret,
    pg_catalog.convert_from(
      vault._crypto_aead_det_decrypt(
        message    => pg_catalog.decode(s.secret, 'base64'),
        additional => pg_catalog.convert_to(s.id::text, 'utf8'),
        key_id     => 0::bigint,
        context    => E'\\x7067736f6469756d'::bytea, -- 'pgsodium'
        nonce      => s.nonce
      ),
      'utf8'
    ) AS decrypted_secret,
    s.created_at,
    s.updated_at
  FROM vault.secrets AS s;
END;
$function$
;


