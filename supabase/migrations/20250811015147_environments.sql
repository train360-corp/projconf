create table "public"."environments" (
    "id" uuid not null default gen_random_uuid(),
    "created_at" timestamp with time zone not null default (now() AT TIME ZONE 'utc'::text),
    "display" text not null default ''::text,
    "project_id" uuid not null,
    unique (display, project_id),
    check ( length(trim(display)) > 1 AND display ~ '^[[:alnum:] _]+$' )
);


alter table "public"."environments" enable row level security;

CREATE UNIQUE INDEX environments_pkey ON public.environments USING btree (id);

alter table "public"."environments" add constraint "environments_pkey" PRIMARY KEY using index "environments_pkey";

alter table "public"."environments" add constraint "environments_project_id_fkey" FOREIGN KEY (project_id) REFERENCES projects(id) ON UPDATE CASCADE ON DELETE CASCADE not valid;

alter table "public"."environments" validate constraint "environments_project_id_fkey";

grant delete on table "public"."environments" to "anon";

grant insert on table "public"."environments" to "anon";

grant references on table "public"."environments" to "anon";

grant select on table "public"."environments" to "anon";

grant trigger on table "public"."environments" to "anon";

grant truncate on table "public"."environments" to "anon";

grant update on table "public"."environments" to "anon";

grant delete on table "public"."environments" to "authenticated";

grant insert on table "public"."environments" to "authenticated";

grant references on table "public"."environments" to "authenticated";

grant select on table "public"."environments" to "authenticated";

grant trigger on table "public"."environments" to "authenticated";

grant truncate on table "public"."environments" to "authenticated";

grant update on table "public"."environments" to "authenticated";

grant delete on table "public"."environments" to "service_role";

grant insert on table "public"."environments" to "service_role";

grant references on table "public"."environments" to "service_role";

grant select on table "public"."environments" to "service_role";

grant trigger on table "public"."environments" to "service_role";

grant truncate on table "public"."environments" to "service_role";

grant update on table "public"."environments" to "service_role";


