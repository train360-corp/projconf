create type "public"."generator" as enum ('STATIC', 'RANDOM');

create extension if not exists "pg_jsonschema" with schema "extensions";

CREATE OR REPLACE FUNCTION is_valid_generator_data(
  gtype generator,
  gdata jsonb
) RETURNS boolean
LANGUAGE sql
IMMUTABLE
AS $$
SELECT CASE gtype
  WHEN 'STATIC' THEN
    jsonb_matches_schema(
      schema := '{"type": "string"}'::json,
      instance := gdata
    )
  WHEN 'RANDOM' THEN
    jsonb_matches_schema(
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
  END;
$$;

