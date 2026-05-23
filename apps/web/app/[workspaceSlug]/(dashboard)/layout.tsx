"use client";

import { DashboardLayout } from "@multica/views/layout";
import { CimeriaIcon } from "@multica/ui/components/common/cimeria-icon";
import { SearchCommand, SearchTrigger } from "@multica/views/search";
import { ChatFab, ChatWindow } from "@multica/views/chat";

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <DashboardLayout
      loadingIndicator={<CimeriaIcon className="size-6" />}
      searchSlot={<SearchTrigger />}
      extra={<><SearchCommand /><ChatWindow /><ChatFab /></>}
    >
      {children}
    </DashboardLayout>
  );
}
