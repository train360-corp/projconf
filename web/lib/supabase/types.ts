import { Database } from "@/lib/supabase/types.gen";



export interface SupabaseRuntimeEnvironmentVariables {
  PROJCONF_URL: string;
  X_ADMIN_API_KEY: string;
}

export type WithSupabaseEnv<T extends object = object> = T & {
  supabase: SupabaseRuntimeEnvironmentVariables
}
export type RuntimeEnvironmentVariable = keyof WithSupabaseEnv["supabase"];
const envObj: { [Key in RuntimeEnvironmentVariable]: Key } = {
  X_ADMIN_API_KEY: "X_ADMIN_API_KEY",
  PROJCONF_URL: "PROJCONF_URL",
};
export const ENV_KEYS: RuntimeEnvironmentVariable[] = Object.keys(envObj) as RuntimeEnvironmentVariable[];


export type Table = string & keyof Database["public"]["Tables"]