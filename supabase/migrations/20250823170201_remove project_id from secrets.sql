alter table "public"."secrets" drop constraint "secrets_project_id_fkey";

alter table "public"."secrets" drop constraint "secrets_project_id_variable_id_environment_id_key";

drop index if exists "public"."secrets_project_id_variable_id_environment_id_key";

alter table "public"."secrets" drop column "project_id";


