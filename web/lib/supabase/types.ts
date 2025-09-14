import { Database } from "@/lib/supabase/types.gen";


export interface SupabaseRuntimeEnvironmentVariables {
  SUPABASE_FRONTEND_URL: string
  SUPABASE_BACKEND_URL: string
  SUPABASE_PUBLISHABLE_OR_ANON_KEY: string
  X_ADMIN_API_KEY: string
}

export type WithSupabaseEnv<T extends object = object> = T & {
  supabase: SupabaseRuntimeEnvironmentVariables
}
export type RuntimeEnvironmentVariable = keyof WithSupabaseEnv["supabase"];
const envObj: { [Key in RuntimeEnvironmentVariable]: Key } = {
  SUPABASE_FRONTEND_URL: "SUPABASE_FRONTEND_URL",
  SUPABASE_PUBLISHABLE_OR_ANON_KEY: "SUPABASE_PUBLISHABLE_OR_ANON_KEY",
  SUPABASE_BACKEND_URL: "SUPABASE_BACKEND_URL",
  X_ADMIN_API_KEY: "X_ADMIN_API_KEY",
};
export const ENV_KEYS: RuntimeEnvironmentVariable[] = Object.keys(envObj) as RuntimeEnvironmentVariable[];


export type Table = string & keyof Database["public"]["Tables"]