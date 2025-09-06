"use client";

import * as React from "react";
import { Dispatch, SetStateAction, useState } from "react";
import { ChevronsUpDown, Plus } from "lucide-react";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { SidebarMenu, SidebarMenuButton, SidebarMenuItem, useSidebar, } from "@/components/ui/sidebar";
import { Tables } from "@/lib/supabase/types.gen";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { WithSupabaseEnv } from "@/lib/supabase/types";
import { createClient } from "@/lib/supabase/client";
import { toast } from "sonner";
import { useRouter } from "next/navigation";



const schema = z.object({
  display: z
    .string()
    .trim()
    .min(2, { message: "Must be at least 2 characters" })
    .regex(/^[\p{L}\p{N}_ ]+$/u, {
      message: "Only letters, numbers, spaces, and underscores allowed",
    }),
});

type FormValues = z.infer<typeof schema>


export function ProjectSwitcher({
                                  projects,
                                  project,
                                  ...props
                                }: WithSupabaseEnv<{
  projects: readonly Tables<"projects">[]
  setProjectsAction: Dispatch<SetStateAction<readonly Tables<"projects">[]>>
  project: Tables<"projects"> | null
}>) {

  const router = useRouter();
  const { isMobile } = useSidebar();
  const [ open, setOpen ] = useState<boolean>(false);
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { display: "" },
  });

  return (
    <SidebarMenu>
      <Dialog open={open} onOpenChange={(open) => {
        setOpen(open);
        if (!open) form.reset();
      }}>
        <SidebarMenuItem>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <SidebarMenuButton
                size="lg"
                className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
              >
                {project === null ? (
                  <div className="grid flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-medium">No project selected</span>
                    <span className="truncate text-xs text-muted">
                    Select a project to continue
                  </span>
                  </div>
                ) : (
                  <>
                    <Avatar>
                      <AvatarFallback>{project.display.trim()[0]}</AvatarFallback>
                    </Avatar>
                    <div className="grid flex-1 text-left text-sm leading-tight">
                      <span className="truncate font-medium">{project.display}</span>
                      <span className="truncate text-xs text-muted">{project.id}</span>
                    </div>
                  </>
                )}
                <ChevronsUpDown className="ml-auto"/>
              </SidebarMenuButton>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
              align="start"
              side={isMobile ? "bottom" : "right"}
              sideOffset={4}
            >
              <DropdownMenuLabel className="text-muted-foreground text-xs">
                Projects
              </DropdownMenuLabel>
              {projects.map((p, index) => (
                <DropdownMenuItem
                  key={p.id}
                  onClick={() => router.push(`/projects/${p.id}`)}
                  className="gap-2 p-2"
                >
                  <Avatar>
                    <AvatarFallback>{p.display.trim()[0]}</AvatarFallback>
                  </Avatar>
                  {p.display}
                  <DropdownMenuShortcut>⌘{index + 1}</DropdownMenuShortcut>
                </DropdownMenuItem>
              ))}
              <DropdownMenuSeparator/>

              <DialogTrigger asChild>
                <DropdownMenuItem className="gap-2 p-2 cursor-pointer">
                  <div className="flex size-6 items-center justify-center rounded-md border bg-transparent">
                    <Plus className="size-4"/>
                  </div>
                  <div className="text-muted-foreground font-medium">
                    Create project
                  </div>
                </DropdownMenuItem>
              </DialogTrigger>

            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarMenuItem>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Project</DialogTitle>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(async ({ display }) => {

              const supabase = createClient(props);
              const project = await supabase.from("projects").insert({
                display: display
              }).select().single();

              if (project.error === null) {
                props.setProjectsAction(projects => [ ...projects, project.data ]);
                router.push(`/projects/${project.data.id}`);
                setOpen(false);
              } else {
                console.error(project.error);
                if (project.error.code === "23505") {
                  toast.error("Project already exists", {
                    description: `a project with display "${display}" already exists`,
                  });
                } else toast.error("Unexpected Error", {
                  description: project.error.message,
                });
              }
            })} className="space-y-4">
              <FormField
                control={form.control}
                name="display"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Display</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="My Project"
                        disabled={form.formState.isSubmitting}
                        {...field}
                      />
                    </FormControl>
                    <FormMessage/>
                  </FormItem>
                )}
              />
              <DialogFooter>
                <Button type="submit">Create</Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </SidebarMenu>
  );
}