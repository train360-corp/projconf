alter table "public"."clients" add column "project_id" uuid not null;

alter table "public"."clients" add constraint "clients_project_id_fkey" FOREIGN KEY (project_id) REFERENCES projects(id) ON UPDATE CASCADE ON DELETE CASCADE not valid;

alter table "public"."clients" validate constraint "clients_project_id_fkey";


