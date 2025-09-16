drop trigger if exists "environments_after_actions" on "public"."environments";

drop trigger if exists "variables_after_actions" on "public"."variables";

drop policy "select based on RLS on clients_secrets" on "public"."clients";

drop policy "x-admin-api-key" on "public"."clients";

drop policy "select based on request headers" on "public"."clients_secrets";

drop policy "x-admin-api-key" on "public"."clients_secrets";

drop policy "select based on RLS on projects" on "public"."environments";

drop policy "x-admin-api-key" on "public"."environments";

drop policy "select based on RLS on clients" on "public"."projects";

drop policy "x-admin-api-key" on "public"."projects";

drop policy "select based on RLS on environments" on "public"."secrets";

drop policy "x-admin-api-key" on "public"."secrets";

drop policy "select based on RLS on projects" on "public"."variables";

drop policy "x-admin-api-key" on "public"."variables";

alter table "public"."variables" drop constraint "variables_generator_data_check";

drop function if exists "public"."create_client_and_secret"(p_display text, p_env_id uuid);

drop function if exists "public"."create_client_secret"(p_client_id uuid, p_secret text);

drop function if exists "public"."environments_after_actions"();

drop function if exists "public"."get_default_secret"(variable_id uuid);

drop function if exists "public"."is_admin_client"();

drop function if exists "public"."is_uuid"(s text);

drop function if exists "public"."is_valid_generator_data"(gtype public.generator, gdata jsonb);

drop function if exists "public"."variables_after_actions"();

drop function if exists "public"."verify_client_secret"(p_secret_id uuid, p_secret text);

CREATE OR REPLACE FUNCTION private.is_valid_generator_data(gtype public.generator, gdata jsonb)
    RETURNS boolean
    LANGUAGE sql
    IMMUTABLE
    SET search_path TO ''
AS $function$SELECT CASE gtype
                        WHEN 'STATIC' THEN
                            extensions.jsonb_matches_schema(
                                    schema := '{"type": "string"}'::json,
                                    instance := gdata
                            )
                        WHEN 'RANDOM' THEN
                            extensions.jsonb_matches_schema(
                                    schema := '{
                                      "type": "object",
                                      "properties": {
                                        "length": { "type": "number" },
                                        "letters": { "type": "boolean" },
                                        "numbers": { "type": "boolean" },
                                        "symbols": { "type": "boolean" }
                                      },
                                      "required": ["length", "letters", "numbers", "symbols"],
                                      "additionalProperties": false
                                    }'::json,
                                    instance := gdata
                            )
                        ELSE
                            false
                        END;$function$
;

set check_function_bodies = off;

CREATE OR REPLACE FUNCTION private.is_admin_client()
    RETURNS boolean
    LANGUAGE plpgsql
    SECURITY DEFINER
    SET search_path TO ''
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

CREATE OR REPLACE FUNCTION private.create_client_and_secret(p_display text, p_env_id uuid)
    RETURNS TABLE(client_id uuid, secret_id uuid, secret text)
    LANGUAGE plpgsql
    SECURITY DEFINER
    SET search_path TO ''
AS $function$BEGIN

    -- must be admin
    IF NOT (private.is_admin_client()) THEN
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

CREATE OR REPLACE FUNCTION private.create_client_secret(p_client_id uuid, p_secret text)
    RETURNS uuid
    LANGUAGE plpgsql
    SECURITY DEFINER
    SET search_path TO ''
AS $function$DECLARE
    secret_id uuid;
BEGIN

    -- must be admin
    IF NOT (private.is_admin_client()) THEN
        RAISE EXCEPTION 'unauthorized';
    END IF;

    INSERT INTO public.clients_secrets (hash, client_id)
    VALUES (extensions.crypt(p_secret, extensions.gen_salt('bf', 12)), -- bcrypt cost 12
            p_client_id)
    RETURNING id INTO secret_id;
    RETURN secret_id;
END;$function$
;

CREATE OR REPLACE FUNCTION private.environments_after_actions()
    RETURNS trigger
    LANGUAGE plpgsql
    SECURITY DEFINER
    SET search_path TO ''
AS $function$declare
    var public.variables%rowtype;
begin

    if tg_op = 'INSERT' then
        for var in (select * from public.variables where project_id = new.project_id)
            loop
                insert into public.secrets(variable_id, environment_id)
                values (var.id, new.id);
            end loop;
    end if;

    return coalesce(new, old);

end;$function$
;

CREATE OR REPLACE FUNCTION private.get_default_secret(variable_id uuid)
    RETURNS text
    LANGUAGE plpgsql
    SET search_path TO ''
AS $function$
declare
    -- shared
    variable public.variables%rowtype;
    val      text := '';

    -- RANDOM
    len      int;
    charset  text := '';
begin

    select * from public.variables v where v.id = variable_id limit 1 into variable;
    if variable.id is null then
        raise exception 'variable (id=%) not found', variable_id;
    end if;

    case variable.generator_type
        when 'STATIC'::public.generator then val := variable.generator_data #>> '{}';
        when 'RANDOM'::public.generator then
            IF (variable.generator_data ->> 'letters')::boolean THEN
                charset := charset || 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ';
            END IF;
            IF (variable.generator_data ->> 'numbers')::boolean THEN
                charset := charset || '0123456789';
            END IF;
            IF (variable.generator_data ->> 'symbols')::boolean THEN
                charset := charset || '!@#$%^&*()-_=+[]{};:,.<>?';
            END IF;

            IF charset = '' THEN
                RAISE EXCEPTION 'charset empty!';
            END IF;

            len := (variable.generator_data->>'length')::int;

            FOR i IN 1..len LOOP
                    val := val || substr(charset, floor(random() * length(charset) + 1)::int, 1);
                END LOOP;


        else raise exception 'unhandled generator type (%)', variable.generator_type;
        end case;

    if val is null then
        raise exception 'val unexpectedly null (generator=%)', variable.generator_type;
    end if;

    return val;
end;
$function$
;

CREATE OR REPLACE FUNCTION private.is_uuid(s text)
    RETURNS boolean
    LANGUAGE plpgsql
    IMMUTABLE
    SET search_path TO ''
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

CREATE OR REPLACE FUNCTION private.variables_after_actions()
    RETURNS trigger
    LANGUAGE plpgsql
    SECURITY DEFINER
    SET search_path TO ''
AS $function$declare
    env public.environments%rowtype;
begin

    if tg_op = 'INSERT' then
        for env in (select * from public.environments where project_id = new.project_id)
            loop
                insert into public.secrets(variable_id, environment_id)
                values (new.id, env.id);
            end loop;
    end if;

    return coalesce(new, old);

end;$function$
;

CREATE OR REPLACE FUNCTION private.verify_client_secret(p_secret_id uuid, p_secret text)
    RETURNS boolean
    LANGUAGE sql
    STABLE SECURITY DEFINER
    SET search_path TO ''
AS $function$
SELECT EXISTS (
    SELECT 1
    FROM public.clients_secrets cs
    WHERE cs.id = p_secret_id
      AND cs.hash = crypt(p_secret, cs.hash)
);
$function$
;


alter table "public"."variables" add constraint "variables_generator_data_check" CHECK (private.is_valid_generator_data(generator_type, generator_data)) not valid;

alter table "public"."variables" validate constraint "variables_generator_data_check";

create policy "select based on RLS on clients_secrets"
on "public"."clients"
as permissive
for select
to anon
using ((( SELECT private.is_admin_client() AS is_admin_client) OR (EXISTS ( SELECT 1
   FROM clients_secrets cs
  WHERE (cs.client_id = clients.id)))));


create policy "x-admin-api-key"
on "public"."clients"
as permissive
for all
to anon
using (( SELECT private.is_admin_client() AS is_admin_client));


create policy "select based on request headers"
on "public"."clients_secrets"
as permissive
for select
to anon
using ((( SELECT private.is_uuid(((current_setting('request.headers'::text, true))::json ->> 'x-client-secret-id'::text)) AS is_uuid) AND ((((current_setting('request.headers'::text, true))::json ->> 'x-client-secret-id'::text))::uuid = id) AND private.verify_client_secret((((current_setting('request.headers'::text, true))::json ->> 'x-client-secret-id'::text))::uuid, ((current_setting('request.headers'::text, true))::json ->> 'x-client-secret'::text))));


create policy "x-admin-api-key"
on "public"."clients_secrets"
as permissive
for all
to anon
using (( SELECT private.is_admin_client() AS is_admin_client));


create policy "select based on RLS on projects"
on "public"."environments"
as permissive
for select
to anon
using ((( SELECT private.is_admin_client() AS is_admin_client) OR (EXISTS ( SELECT 1
   FROM clients c
  WHERE (c.environment_id = environments.id)))));


create policy "x-admin-api-key"
on "public"."environments"
as permissive
for all
to anon
using (( SELECT private.is_admin_client() AS is_admin_client));


create policy "select based on RLS on clients"
on "public"."projects"
as permissive
for select
to anon
using ((( SELECT private.is_admin_client() AS is_admin_client) OR (EXISTS ( SELECT 1
   FROM environments ev
  WHERE (ev.project_id = projects.id)))));


create policy "x-admin-api-key"
on "public"."projects"
as permissive
for all
to anon
using (( SELECT private.is_admin_client() AS is_admin_client));


create policy "select based on RLS on environments"
on "public"."secrets"
as permissive
for select
to anon
using ((( SELECT private.is_admin_client() AS is_admin_client) OR (EXISTS ( SELECT 1
   FROM environments e
  WHERE (e.id = secrets.environment_id)))));


create policy "x-admin-api-key"
on "public"."secrets"
as permissive
for all
to anon
using (( SELECT private.is_admin_client() AS is_admin_client));


create policy "select based on RLS on projects"
on "public"."variables"
as permissive
for select
to anon
using ((( SELECT private.is_admin_client() AS is_admin_client) OR (EXISTS ( SELECT 1
   FROM projects p
  WHERE (p.id = variables.project_id)))));


create policy "x-admin-api-key"
on "public"."variables"
as permissive
for all
to anon
using (( SELECT private.is_admin_client() AS is_admin_client));


CREATE TRIGGER environments_after_actions AFTER INSERT OR DELETE OR UPDATE ON public.environments FOR EACH ROW EXECUTE FUNCTION private.environments_after_actions();

CREATE TRIGGER variables_after_actions AFTER INSERT OR DELETE OR UPDATE ON public.variables FOR EACH ROW EXECUTE FUNCTION private.variables_after_actions();




