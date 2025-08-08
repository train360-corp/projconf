create table "public"."projects" (
    "id" uuid not null default gen_random_uuid(),
    "display" text not null default ''::text
);


alter table "public"."projects" enable row level security;

create table "public"."variables" (
    "id" uuid not null default gen_random_uuid(),
    "description" text not null default ''::text,
    "key" text not null
);


alter table "public"."variables" enable row level security;

CREATE UNIQUE INDEX projects_pkey ON public.projects USING btree (id);

CREATE UNIQUE INDEX variables_key_key ON public.variables USING btree (key);

CREATE UNIQUE INDEX variables_pkey ON public.variables USING btree (id, key);

alter table "public"."projects" add constraint "projects_pkey" PRIMARY KEY using index "projects_pkey";

alter table "public"."variables" add constraint "variables_pkey" PRIMARY KEY using index "variables_pkey";

alter table "public"."variables" add constraint "variables_key_check" CHECK ((key ~ '^[A-Z_][A-Z0-9_]*$'::text)) not valid;

alter table "public"."variables" validate constraint "variables_key_check";

alter table "public"."variables" add constraint "variables_key_key" UNIQUE using index "variables_key_key";

grant delete on table "public"."projects" to "anon";

grant insert on table "public"."projects" to "anon";

grant references on table "public"."projects" to "anon";

grant select on table "public"."projects" to "anon";

grant trigger on table "public"."projects" to "anon";

grant truncate on table "public"."projects" to "anon";

grant update on table "public"."projects" to "anon";

grant delete on table "public"."projects" to "authenticated";

grant insert on table "public"."projects" to "authenticated";

grant references on table "public"."projects" to "authenticated";

grant select on table "public"."projects" to "authenticated";

grant trigger on table "public"."projects" to "authenticated";

grant truncate on table "public"."projects" to "authenticated";

grant update on table "public"."projects" to "authenticated";

grant delete on table "public"."projects" to "service_role";

grant insert on table "public"."projects" to "service_role";

grant references on table "public"."projects" to "service_role";

grant select on table "public"."projects" to "service_role";

grant trigger on table "public"."projects" to "service_role";

grant truncate on table "public"."projects" to "service_role";

grant update on table "public"."projects" to "service_role";

grant delete on table "public"."variables" to "anon";

grant insert on table "public"."variables" to "anon";

grant references on table "public"."variables" to "anon";

grant select on table "public"."variables" to "anon";

grant trigger on table "public"."variables" to "anon";

grant truncate on table "public"."variables" to "anon";

grant update on table "public"."variables" to "anon";

grant delete on table "public"."variables" to "authenticated";

grant insert on table "public"."variables" to "authenticated";

grant references on table "public"."variables" to "authenticated";

grant select on table "public"."variables" to "authenticated";

grant trigger on table "public"."variables" to "authenticated";

grant truncate on table "public"."variables" to "authenticated";

grant update on table "public"."variables" to "authenticated";

grant delete on table "public"."variables" to "service_role";

grant insert on table "public"."variables" to "service_role";

grant references on table "public"."variables" to "service_role";

grant select on table "public"."variables" to "service_role";

grant trigger on table "public"."variables" to "service_role";

grant truncate on table "public"."variables" to "service_role";

grant update on table "public"."variables" to "service_role";


