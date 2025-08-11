create table "public"."clients_secrets" (
    "id" uuid not null default gen_random_uuid(),
    "created_at" timestamp with time zone not null default (now() AT TIME ZONE 'utc'::text),
    "hash" text not null,
    "client_id" uuid not null
);


alter table "public"."clients_secrets" enable row level security;

CREATE UNIQUE INDEX clients_secrets_pkey ON public.clients_secrets USING btree (id);

alter table "public"."clients_secrets" add constraint "clients_secrets_pkey" PRIMARY KEY using index "clients_secrets_pkey";

alter table "public"."clients_secrets" add constraint "clients_secrets_client_id_fkey" FOREIGN KEY (client_id) REFERENCES clients(id) ON UPDATE CASCADE ON DELETE CASCADE not valid;

alter table "public"."clients_secrets" validate constraint "clients_secrets_client_id_fkey";

alter table "public"."clients_secrets" add constraint "clients_secrets_hash_check" CHECK ((hash ~ '^\$(2[aby])\$12\$[./A-Za-z0-9]{53}$'::text)) not valid;

alter table "public"."clients_secrets" validate constraint "clients_secrets_hash_check";

set check_function_bodies = off;

CREATE OR REPLACE FUNCTION public.verify_client_secret(p_secret_id uuid, p_secret text)
 RETURNS boolean
 LANGUAGE sql
 STABLE SECURITY DEFINER
 SET search_path TO 'public', 'extensions'
AS $function$
SELECT EXISTS (
    SELECT 1
    FROM public.clients_secrets cs
    WHERE cs.id = p_secret_id
      AND cs.hash = crypt(p_secret, cs.hash)
);
$function$
;

grant delete on table "public"."clients_secrets" to "anon";

grant insert on table "public"."clients_secrets" to "anon";

grant references on table "public"."clients_secrets" to "anon";

grant select on table "public"."clients_secrets" to "anon";

grant trigger on table "public"."clients_secrets" to "anon";

grant truncate on table "public"."clients_secrets" to "anon";

grant update on table "public"."clients_secrets" to "anon";

grant delete on table "public"."clients_secrets" to "authenticated";

grant insert on table "public"."clients_secrets" to "authenticated";

grant references on table "public"."clients_secrets" to "authenticated";

grant select on table "public"."clients_secrets" to "authenticated";

grant trigger on table "public"."clients_secrets" to "authenticated";

grant truncate on table "public"."clients_secrets" to "authenticated";

grant update on table "public"."clients_secrets" to "authenticated";

grant delete on table "public"."clients_secrets" to "service_role";

grant insert on table "public"."clients_secrets" to "service_role";

grant references on table "public"."clients_secrets" to "service_role";

grant select on table "public"."clients_secrets" to "service_role";

grant trigger on table "public"."clients_secrets" to "service_role";

grant truncate on table "public"."clients_secrets" to "service_role";

grant update on table "public"."clients_secrets" to "service_role";


