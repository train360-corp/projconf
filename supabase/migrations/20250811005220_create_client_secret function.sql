set check_function_bodies = off;

CREATE OR REPLACE FUNCTION public.create_client_secret(p_client_id uuid)
 RETURNS TABLE(secret_id uuid, secret text)
 LANGUAGE plpgsql
 SET search_path TO 'public', 'extensions'
AS $function$
DECLARE
  v_secret text;
BEGIN
  -- generate a random 32-char secret in bcrypt-safe charset
  v_secret := encode(gen_random_bytes(24), 'base64'); -- ~32 chars, base64-safe
  
  -- insert hashed secret
  INSERT INTO public.clients_secrets (hash, client_id)
  VALUES (
    crypt(v_secret, gen_salt('bf', 12)), -- bcrypt cost 12
    p_client_id
  )
  RETURNING id INTO secret_id;

  -- return the unhashed secret to caller
  secret := v_secret;
  RETURN NEXT;
END;
$function$
;


