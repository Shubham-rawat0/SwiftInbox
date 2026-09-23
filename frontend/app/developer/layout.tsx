import {AppSidebar} from "@/components/layout/Sidebar";
import { DeveloperProfileProvider } from "@/components/layout/DeveloperProfileContext";
import { DeveloperSessionGate } from "@/components/developer/DeveloperSessionGate";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";


export default function DeveloperLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <DeveloperProfileProvider>
      <SidebarProvider className="min-h-[calc(100svh-4rem)]">
        <AppSidebar />
        <SidebarInset>
          <main className="flex flex-1 flex-col">
            <DeveloperSessionGate>{children}</DeveloperSessionGate>
          </main>
        </SidebarInset>
      </SidebarProvider>
    </DeveloperProfileProvider>
  );
}