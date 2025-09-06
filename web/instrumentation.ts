/**
 * called once on app start
 */
export function register() {

  if (!process.env.SUPABASE_URL) {
    throw new Error("environment variable 'SUPABASE_URL' is falsy");
  }

  if (!process.env.SUPABASE_PUBLISHABLE_OR_ANON_KEY) {
    throw new Error("environment variable 'SUPABASE_PUBLISHABLE_OR_ANON_KEY' is falsy");
  }

  if (!process.env.X_ADMIN_API_KEY) {
    throw new Error("environment variable 'X_ADMIN_API_KEY' is falsy");
  }

}