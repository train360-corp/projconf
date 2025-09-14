export const getSupabase = () => ({
  SUPABASE_URL: process.env.SUPABASE_URL,
  SUPABASE_PUBLISHABLE_OR_ANON_KEY: process.env.SUPABASE_PUBLISHABLE_OR_ANON_KEY,
  X_ADMIN_API_KEY: process.env.X_ADMIN_API_KEY,
})