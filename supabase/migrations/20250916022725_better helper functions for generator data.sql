set check_function_bodies = off;

CREATE OR REPLACE FUNCTION private.is_valid_random_generator_data(gtype public.generator, gdata jsonb)
 RETURNS boolean
 LANGUAGE sql
 IMMUTABLE
 SET search_path TO ''
AS $function$SELECT extensions.jsonb_matches_schema(
      schema := '{
        "type": "object",
        "properties": {
          "length": { "type": "number" },
          "letters": { "type": "boolean" },
          "numbers": { "type": "boolean" },
          "symbols": { "type": "boolean" }
        },
        "required": ["length", "letters", "numbers", "symbols"],
        "additionalProperties": false
      }'::json,
      instance := gdata
    );$function$
;

CREATE OR REPLACE FUNCTION private.is_valid_static_generator_data(gtype public.generator, gdata jsonb)
 RETURNS boolean
 LANGUAGE sql
 IMMUTABLE
 SET search_path TO ''
AS $function$SELECT extensions.jsonb_matches_schema(
      schema := '{
        "type": "object",
        "properties": {
          "secret-id": {
            "type": "string",
            "format": "uuid"
          }
        },
        "required": [
          "secret-id"
        ],
        "additionalProperties": false
      }'::json,
      instance := gdata
    );$function$
;

CREATE OR REPLACE FUNCTION private.is_valid_generator_data(gtype public.generator, gdata jsonb)
 RETURNS boolean
 LANGUAGE sql
 IMMUTABLE
 SET search_path TO ''
AS $function$SELECT CASE gtype
  WHEN 'STATIC' THEN
    private.is_valid_static_generator_data(gtype, gdata)
  WHEN 'RANDOM' THEN
    private.is_valid_random_generator_data(gtype, gdata)
  ELSE
    false
  END;$function$
;
