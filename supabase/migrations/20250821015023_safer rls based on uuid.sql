drop policy "select based on request headers" on "public"."clients_secrets";

set check_function_bodies = off;

CREATE OR REPLACE FUNCTION public.is_uuid(s text)
 RETURNS boolean
 LANGUAGE plpgsql
 IMMUTABLE
AS $function$
begin
  -- NULL is not a UUID
  if s is null then
    return false;
  end if;

  -- Regex match for canonical UUID format
  -- 8-4-4-4-12 hex characters, version 1–5, variant 8–b
  if s ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' then
    return true;
  else
    return false;
  end if;
end;
$function$
;

create policy "select based on request headers"
on "public"."clients_secrets"
as permissive
for select
to anon
using ((( SELECT is_uuid(((current_setting('request.headers'::text, true))::json ->> 'x-client-secret-id'::text)) AS is_uuid) AND ((((current_setting('request.headers'::text, true))::json ->> 'x-client-secret-id'::text))::uuid = id) AND verify_client_secret((((current_setting('request.headers'::text, true))::json ->> 'x-client-secret-id'::text))::uuid, ((current_setting('request.headers'::text, true))::json ->> 'x-client-secret'::text))));



