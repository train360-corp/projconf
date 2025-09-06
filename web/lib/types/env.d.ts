declare namespace NodeJS {
  interface ProcessEnv {
    SUPABASE_URL: string;
    SUPABASE_PUBLISHABLE_OR_ANON_KEY: string;
    X_ADMIN_API_KEY: string;
  }
}