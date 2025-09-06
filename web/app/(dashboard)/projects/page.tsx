import LayoutDashboard from "@/components/nav/layout-dashboard";



export default async function Page() {
  return (
    <LayoutDashboard projectId={undefined}>
      <div className={"w-full h-full flex flex-col items-center justify-center"}>
        <p className={"text-muted"}>{"No Project Selected"}</p>
      </div>
    </LayoutDashboard>
  )
}