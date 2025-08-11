create table "public"."secrets" (
    "id" uuid not null default gen_random_uuid(),
    "created_at" timestamp with time zone not null default (now() AT TIME ZONE 'utc'::text),
    "project_id" uuid not null,
    "variable_id" uuid not null,
    "environment_id" uuid not null,
    "value" text not null default (''),
    unique (project_id, variable_id, environment_id)
);


alter table "public"."secrets" enable row level security;

CREATE UNIQUE INDEX secrets_pkey ON public.secrets USING btree (id);

alter table "public"."secrets" add constraint "secrets_pkey" PRIMARY KEY using index "secrets_pkey";

alter table "public"."secrets" add constraint "secrets_environment_id_fkey" FOREIGN KEY (environment_id) REFERENCES environments(id) ON UPDATE CASCADE ON DELETE CASCADE not valid;

alter table "public"."secrets" validate constraint "secrets_environment_id_fkey";

alter table "public"."secrets" add constraint "secrets_project_id_fkey" FOREIGN KEY (project_id) REFERENCES projects(id) ON UPDATE CASCADE ON DELETE CASCADE not valid;

alter table "public"."secrets" validate constraint "secrets_project_id_fkey";

alter table "public"."secrets" add constraint "secrets_variable_id_fkey" FOREIGN KEY (variable_id) REFERENCES variables(id) ON UPDATE CASCADE ON DELETE CASCADE not valid;

alter table "public"."secrets" validate constraint "secrets_variable_id_fkey";

grant delete on table "public"."secrets" to "anon";

grant insert on table "public"."secrets" to "anon";

grant references on table "public"."secrets" to "anon";

grant select on table "public"."secrets" to "anon";

grant trigger on table "public"."secrets" to "anon";

grant truncate on table "public"."secrets" to "anon";

grant update on table "public"."secrets" to "anon";

grant delete on table "public"."secrets" to "authenticated";

grant insert on table "public"."secrets" to "authenticated";

grant references on table "public"."secrets" to "authenticated";

grant select on table "public"."secrets" to "authenticated";

grant trigger on table "public"."secrets" to "authenticated";

grant truncate on table "public"."secrets" to "authenticated";

grant update on table "public"."secrets" to "authenticated";

grant delete on table "public"."secrets" to "service_role";

grant insert on table "public"."secrets" to "service_role";

grant references on table "public"."secrets" to "service_role";

grant select on table "public"."secrets" to "service_role";

grant trigger on table "public"."secrets" to "service_role";

grant truncate on table "public"."secrets" to "service_role";

grant update on table "public"."secrets" to "service_role";


