set check_function_bodies = off;

CREATE OR REPLACE FUNCTION public.is_admin_client()
 RETURNS boolean
 LANGUAGE plpgsql
 SECURITY DEFINER
AS $function$declare
  hdr text := coalesce(((current_setting('request.headers', true))::json ->> 'x-admin-api-key'), '');
  guc text := coalesce(current_setting('projconf.x_admin_api_key', true), '');

  a bytea;
  b bytea;
  diff int := 0;
  i int;
begin

  -- if one is not set, should not authenticate
  if guc = '' then
    raise WARNING 'configuration parameter projconf.x_admin_api_key is not set';
    return false;
  end if;

  raise warning 'headers[x-admin-api-key]=%', hdr;

  -- hash both to fixed length
  a := extensions.digest(hdr, 'sha256'::text);
  b := extensions.digest(guc, 'sha256'::text);

  -- constant-time comparison
  for i in 0 .. length(a)-1 loop
    diff := diff | (get_byte(a,i) # get_byte(b,i));
  end loop;

  return diff = 0;
end;$function$
;


