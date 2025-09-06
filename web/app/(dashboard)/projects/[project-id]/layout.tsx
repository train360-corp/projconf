import * as React from "react";
import { ReactNode } from "react";
import LayoutDashboard from "@/components/nav/layout-dashboard";



export default async function Layout(props: Readonly<{
  children: ReactNode;
  params: Promise<{
    ["project-id"]: string;
  }>;
}>) {

  const params = await props.params;

  return (
    <LayoutDashboard projectId={params["project-id"]}>
      {props.children}
    </LayoutDashboard>
  );
};