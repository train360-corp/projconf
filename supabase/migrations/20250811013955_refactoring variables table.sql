alter table "public"."variables" drop constraint "variables_key_key";

alter table "public"."variables" drop constraint "variables_pkey";

drop index if exists "public"."variables_key_key";

drop index if exists "public"."variables_pkey";

alter table "public"."variables" add column "project_id" uuid not null;

CREATE UNIQUE INDEX public_variables_unique_cols_key_project_id ON public.variables USING btree (project_id, key);

CREATE UNIQUE INDEX variables_pkey ON public.variables USING btree (id);

alter table "public"."variables" add constraint "variables_pkey" PRIMARY KEY using index "variables_pkey";

alter table "public"."variables" add constraint "public_variables_unique_cols_key_project_id" UNIQUE using index "public_variables_unique_cols_key_project_id";

alter table "public"."variables" add constraint "variables_project_id_fkey" FOREIGN KEY (project_id) REFERENCES projects(id) ON UPDATE CASCADE ON DELETE CASCADE not valid;

alter table "public"."variables" validate constraint "variables_project_id_fkey";


