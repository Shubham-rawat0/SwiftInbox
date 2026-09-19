import {AppSidebar} from "@/components/layout/Sidebar";
import { DeveloperProfileProvider } from "@/components/layout/DeveloperProfileContext";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";


export default function DeveloperLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <DeveloperProfileProvider>
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          <main>{children}</main>
        </SidebarInset>
      </SidebarProvider>
    </DeveloperProfileProvider>
  );
}