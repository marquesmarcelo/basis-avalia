import { Skeleton } from "@/components/ui/skeleton"

export function AppShellSkeleton() {
  return (
    <div className="flex min-h-screen flex-1 flex-col">
      <header className="flex h-14 shrink-0 items-center gap-3 border-b bg-background px-4 md:px-6">
        <Skeleton className="h-5 w-28" />
        <Skeleton className="h-5 w-16" />
        <div className="ml-auto">
          <Skeleton className="size-8 rounded-full" />
        </div>
      </header>
      <main className="w-full flex-1 space-y-3 px-4 py-6 md:px-6">
        <Skeleton className="h-6 w-64" />
        <Skeleton className="h-40 w-full" />
      </main>
      <footer className="flex h-10 shrink-0 items-center justify-center border-t bg-background px-4">
        <Skeleton className="h-3 w-48" />
      </footer>
    </div>
  )
}
