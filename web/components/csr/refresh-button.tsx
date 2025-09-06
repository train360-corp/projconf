"use client";

import { RefreshCcw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useRouter } from "next/navigation";



export const RefreshButton = () => {
  const router = useRouter()
  return (
    <Button
      type={"button"}
      size={"icon"}
      variant={"ghost"}
      onClick={() => router.refresh()}
    >
      <RefreshCcw/>
    </Button>
  );
}