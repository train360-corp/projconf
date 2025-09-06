import { ReactNode } from "react";



export default function LayoutEnvironments({ children }: {
  environmentId: string | undefined;
  projectId: string | undefined;
  children: ReactNode;
}) {

  return (
    children
  );

}