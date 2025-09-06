import { Database } from "@/lib/supabase/types.gen";

export type WithSupabaseEnv<T extends object = object> = T & {
  supabase: {
    SUPABASE_URL: string
    SUPABASE_PUBLISHABLE_OR_ANON_KEY: string
    X_ADMIN_API_KEY: string
  }
}

export type Table = string & keyof Database["public"]["Tables"]