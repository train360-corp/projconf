alter table "public"."variables" add column "generator_data" jsonb not null;

alter table "public"."variables" add column "generator_type" generator not null;

alter table "public"."variables" add constraint "variables_generator_data_check" CHECK (is_valid_generator_data(generator_type, generator_data)) not valid;

alter table "public"."variables" validate constraint "variables_generator_data_check";


