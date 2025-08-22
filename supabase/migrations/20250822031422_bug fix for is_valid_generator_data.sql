set check_function_bodies = off;

CREATE OR REPLACE FUNCTION public.is_valid_generator_data(gtype generator, gdata jsonb)
 RETURNS boolean
 LANGUAGE sql
 IMMUTABLE
AS $function$SELECT CASE gtype
  WHEN 'STATIC' THEN
    extensions.jsonb_matches_schema(
      schema := '{"type": "string"}'::json,
      instance := gdata
    )
  WHEN 'RANDOM' THEN
    extensions.jsonb_matches_schema(
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
    )
  ELSE
    false
  END;$function$
;


