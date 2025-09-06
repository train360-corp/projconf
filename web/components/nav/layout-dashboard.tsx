import { createClient } from "@/lib/supabase/server";
import { FullPageError } from "@/components/error-handling/error";
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/nav/app-sidebar";
import { Separator } from "@/components/ui/separator";
import { ReactNode } from "react";



export default async function LayoutDashboard({ projectId, children }: {
  projectId: string | undefined;
  children: ReactNode;
}) {

  const supabase = await createClient();

  // load projects
  const { data: projects, error: projectsError } = await supabase.from("projects").select();
  if (projectsError) return (
    <FullPageError
      error={projectsError}
    />
  );

  const project = projects?.find(project => project.id === projectId);
  if (!project && typeof projectId === "string") return (
    <FullPageError
      error={"Project not found"}
      description={`A project with ID \"${projectId}\" was not found.`}
    />
  )

  return (
    <SidebarProvider>
      <AppSidebar
        project={project ?? null}
        initialProjects={projects}
        supabase={{
          SUPABASE_URL: process.env.SUPABASE_URL,
          SUPABASE_PUBLISHABLE_OR_ANON_KEY: process.env.SUPABASE_PUBLISHABLE_OR_ANON_KEY,
          X_ADMIN_API_KEY: process.env.X_ADMIN_API_KEY,
        }}
      />
      <SidebarInset>
        <header
          className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
          <div className="flex items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1"/>
            <Separator
              orientation="vertical"
              className="mr-2 data-[orientation=vertical]:h-4"
            />
          </div>
        </header>
        {children}
      </SidebarInset>
    </SidebarProvider>
  );
}