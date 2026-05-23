"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { RedirectIfAuthenticated } from "@/features/landing/components/redirect-if-authenticated";
import Link from "next/link";

export default function LandingPage() {
  const router = useRouter();

  useEffect(() => {
    const timer = setTimeout(() => {
      router.push("/login");
    }, 3000);
    return () => clearTimeout(timer);
  }, [router]);

  return (
    <>
      <RedirectIfAuthenticated />
      <div className="flex min-h-svh flex-col items-center justify-center gap-8 bg-background px-4">
        <h1 className="text-5xl font-bold tracking-tight">Cimeria</h1>
        <p className="max-w-md text-center text-lg text-muted-foreground">
          Welcome to Cimeria. The AI swarm for sales automation.
        </p>
        <Link
          href="/login"
          className="inline-flex h-11 items-center justify-center rounded-md bg-primary px-8 text-base font-medium text-primary-foreground shadow transition-colors hover:bg-primary/90"
        >
          Log In Now
        </Link>
      </div>
    </>
  );
}