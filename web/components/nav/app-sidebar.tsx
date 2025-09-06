"use client";
import * as React from "react";
import { useEffect, useMemo, useState } from "react";
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar";
import { ProjectSwitcher } from "@/components/nav/project-switcher";
import { Tables } from "@/lib/supabase/types.gen";
import { WithSupabaseEnv } from "@/lib/supabase/types";
import { NavEnvironments } from "@/components/nav/nav-environments";
import { createClient } from "@/lib/supabase/client";
import { toast } from "sonner";
import { Skeleton } from "@/components/ui/skeleton";
import Link from "next/link";



export function AppSidebar(props: WithSupabaseEnv<{
  initialProjects: readonly Tables<"projects">[];
  project: Tables<"projects"> | null;
}>) {

  const [ projects, setProjects ] = useState<readonly Tables<"projects">[]>(props.initialProjects);
  const [ environments, setEnvironments ] = useState<readonly Tables<"environments">[] | undefined>(undefined);
  const supabase = useMemo(() => createClient(props), [ props.supabase.SUPABASE_URL, props.supabase.SUPABASE_PUBLISHABLE_OR_ANON_KEY ]);

  useEffect(() => {
    setEnvironments(undefined); // trigger loading state immediately
    const ac = new AbortController();
    (async () => {
      if (!props.project) return;
      const { data, error } = await supabase
        .from("environments")
        .select()
        .eq("project_id", props.project.id)
        .abortSignal(ac.signal);

      if (ac.signal.aborted) return;

      if (error) {
        toast.error(error.message);
      } else {
        setEnvironments(data);
      }
    })();

    return () => ac.abort();
  }, [ props.project, supabase ]);

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <ProjectSwitcher
          projects={projects}
          setProjectsAction={setProjects}
          project={props.project}
          supabase={props.supabase}
        />
      </SidebarHeader>
      <SidebarContent>
        {props.project !== null && (
          <SidebarGroup className="group-data-[collapsible=icon]:hidden">
            <SidebarGroupLabel>
              {"Project"}
            </SidebarGroupLabel>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton asChild>
                  <Link href={`/projects/${props.project.id}`}>
                    <span>{"Dashboard"}</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroup>
        )}
        {props.project !== null && (
          environments === undefined ? (
            <SidebarGroup className="group-data-[collapsible=icon]:hidden">
              <SidebarGroupLabel>
                <Skeleton className="h-4 w-[150px]"/>
              </SidebarGroupLabel>
              <SidebarMenu>
                <SidebarMenuItem>
                  <Skeleton className="h-[20px] m-2 w-[150px]"/>
                </SidebarMenuItem>
                <SidebarMenuItem>
                  <Skeleton className="h-[20px] m-2 w-[100px]"/>
                </SidebarMenuItem>
                <SidebarMenuItem>
                  <Skeleton className="h-[20px] m-2 w-[125px]"/>
                </SidebarMenuItem>
              </SidebarMenu>
            </SidebarGroup>
          ) : (
            <NavEnvironments
              environments={environments}
              supabase={props.supabase}
            />
          )
        )}
      </SidebarContent>
      {/*<SidebarFooter>*/}
      {/*  <NavUser user={data.user} />*/}
      {/*</SidebarFooter>*/}
      <SidebarRail/>
    </Sidebar>
  );
}
