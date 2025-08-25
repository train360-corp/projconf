alter table "public"."clients" add constraint "clients_display_check" CHECK (((length(TRIM(BOTH FROM display)) > 1) AND (display ~ '^[[:alnum:] _]+$'::text))) not valid;

alter table "public"."clients" validate constraint "clients_display_check";


