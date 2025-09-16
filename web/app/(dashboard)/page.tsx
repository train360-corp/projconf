import * as React from "react";
import { FullPageError } from "@/components/error-handling/error";
import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";



export default async function Page() {
  const supabase = await createClient();


  // load projects
  const { data: projects, error: projectsError } = await supabase.from("projects").select();
  if (projectsError) return (
    <FullPageError
      error={"Failed to Load Project(s)"}
      details={projectsError}
    />
  );

  if (projects && projects.length > 0) {
    return redirect(`/projects/${projects[0].id}`);
  }

  return redirect(`/projects`);
}
