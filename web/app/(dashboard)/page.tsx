import * as React from "react";
import { FullPageError } from "@/components/error-handling/error";
import { redirect } from "next/navigation";
import { createServerClient } from "@/lib/clients/server";



export default async function Page() {
  const projconf = createServerClient();


  // load projects
  const { data: projects, error: projectsError } = await projconf.GET("/v1/projects");
  if (projectsError) return (
    <FullPageError {...projectsError} />
  );

  if (projects && projects.length > 0) {
    return redirect(`/projects/${projects[0].id}`);
  }

  return redirect(`/projects`);
}
