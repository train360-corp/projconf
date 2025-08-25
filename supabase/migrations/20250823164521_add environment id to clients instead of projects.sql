drop policy "select based on RLS on projects" on "public"."environments";

drop policy "select based on RLS on clients" on "public"."projects";

alter table "public"."clients" drop constraint "clients_project_id_fkey";

alter table "public"."clients" drop column "project_id";

alter table "public"."clients" add column "environment_id" uuid not null;

alter table "public"."clients" add constraint "clients_environment_id_fkey" FOREIGN KEY (environment_id) REFERENCES environments(id) ON UPDATE CASCADE ON DELETE CASCADE not valid;

alter table "public"."clients" validate constraint "clients_environment_id_fkey";

create policy "select based on RLS on projects"
on "public"."environments"
as permissive
for select
to anon
using ((EXISTS ( SELECT 1
   FROM clients c
  WHERE (c.environment_id = environments.id))));


create policy "select based on RLS on clients"
on "public"."projects"
as permissive
for select
to anon
using ((( SELECT is_admin_client() AS is_admin_client) OR (EXISTS ( SELECT 1
   FROM environments ev
  WHERE (ev.project_id = projects.id)))));



