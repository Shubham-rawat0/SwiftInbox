"use client";

import { useState } from "react";
import type { SyntheticEvent } from "react";
import { toast } from "sonner";

export default function Home() {
  const [username, setUsername] = useState("");
  const [isCreating, setIsCreating] = useState(false);

  const validateUsername = (input: string): string => {
    const beforeAtSign = input.split("@")[0];

    return beforeAtSign
      .toLowerCase()
      .replace(/[^a-z0-9]/g, "")
      .slice(0, 32);
  };

  const createMailbox = async () => {
    const cleanUsername = username.trim();

    if (!cleanUsername) {
      toast.error("Enter a username first");
      return;
    }

    setIsCreating(true);

    try {
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_BASE || "http://localhost:3001"}/api/mailboxes/custom`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            username: cleanUsername,
          }),
        }
      );

      const data = await response.json().catch(() => ({}));

      if (response.status === 409) {
        toast.error("Username already taken", {
          description:
            "This email address already exists. Please choose another username.",
          duration: 5000,
        });
        return;
      }

      if (response.status === 429) {
        toast.error("Rate limit exceeded", {
          description:
            data.error ||
            "Too many mailboxes created. Please try again after 1 hour.",
          duration: 5000,
        });
        return;
      }

      if (!response.ok) {
        toast.error("Unable to create mailbox", {
          description:
            data.error || "Something went wrong. Please try again.",
        });
        return;
      }

      toast.success("Mailbox created", {
        description: "Your temporary email address is ready.",
      });

      window.location.href = `/mailbox/${encodeURIComponent(cleanUsername)}`;
    } catch (error) {
      console.error("Error creating mailbox:", error);

      toast.error("Connection failed", {
        description: "Could not connect to the mail server.",
      });
    } finally {
      setIsCreating(false);
    }
  };

  const handleSubmit = (e: SyntheticEvent<HTMLFormElement>) => {
    e.preventDefault();
    createMailbox();
  };

  const emailAddress = `${username || "username"}@temp.mail.at`;

  return (
    <div className=" bg-[#f7f7f5] text-[#111] dark:bg-[#0b0c0c] dark:text-white">
    
      <main className="mx-auto flex min-h-[calc(100vh-64px)] max-w-6xl items-center px-5 py-12 sm:px-8 lg:py-20">
        <div className="w-full">
          {/* Hero */}
          <section className="mx-auto max-w-3xl text-center">
            <div className="mb-5 inline-flex items-center gap-2 rounded-full border border-black/10 bg-white px-3.5 py-1.5 text-xs font-medium text-black/60 shadow-sm dark:border-white/10 dark:bg-white/[0.04] dark:text-white/50">
              <span className="h-1.5 w-1.5 rounded-full bg-black dark:bg-white" />
              No signup required
            </div>

            <h1 className="text-4xl font-black tracking-[-0.04em] sm:text-5xl md:text-6xl">
              Your inbox.
              <br />
              <span className="text-black/35 dark:text-white/30">
                Without the noise.
              </span>
            </h1>

            <p className="mx-auto mt-5 max-w-xl text-sm leading-6 text-black/55 dark:text-white/45 sm:text-base">
              Create a temporary email address instantly. Use it for
              newsletters, one-time registrations and services you don't
              want in your personal inbox.
            </p>
          </section>

          {/* Email Generator */}
          <section className="mx-auto mt-10 max-w-2xl">
            <div className="overflow-hidden rounded-3xl border border-black/10 bg-white shadow-[0_20px_70px_rgba(0,0,0,0.08)] dark:border-white/10 dark:bg-[#111313] dark:shadow-black/30">
              <div className="p-5 sm:p-7">
                <div className="mb-5 flex items-center justify-between">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-wider text-black/40 dark:text-white/35">
                      Create address
                    </p>
                    <p className="mt-1 text-sm font-medium">
                      Choose your username
                    </p>
                  </div>

                  <div className="rounded-lg bg-black/[0.04] px-2.5 py-1.5 text-[11px] font-medium text-black/45 dark:bg-white/[0.05] dark:text-white/40">
                    24h inbox
                  </div>
                </div>

                <form onSubmit={handleSubmit}>
                  <div className="rounded-2xl border border-black/10 bg-[#f8f8f7] p-2 transition-all focus-within:border-black/30 focus-within:ring-4 focus-within:ring-black/5 dark:border-white/10 dark:bg-[#0b0c0c] dark:focus-within:border-white/30 dark:focus-within:ring-white/5">
                    <div className="flex items-center">
                      <input
                        id="mail"
                        type="text"
                        autoComplete="off"
                        value={username}
                        placeholder="yourname"
                        disabled={isCreating}
                        onChange={(e) =>
                          setUsername(validateUsername(e.target.value))
                        }
                        className="min-w-0 flex-1 bg-transparent px-4 py-3 text-base font-semibold outline-none placeholder:text-black/25 disabled:cursor-not-allowed disabled:opacity-50 dark:placeholder:text-white/20"
                      />

                      <span className="hidden shrink-0 pr-2 text-sm font-medium text-black/35 sm:block dark:text-white/30">
                        @temp.abhi.at
                      </span>

                      <button
                        type="submit"
                        disabled={isCreating || !username.trim()}
                        className="flex h-11 shrink-0 items-center gap-2 rounded-xl bg-black px-4 text-sm font-semibold text-white transition-all hover:bg-black/80 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-30 dark:bg-white dark:text-black dark:hover:bg-white/85"
                      >
                        {isCreating ? (
                          <>
                            <span className="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white dark:border-black/20 dark:border-t-black" />
                            Creating
                          </>
                        ) : (
                          <>
                            Create
                            <svg
                              width="15"
                              height="15"
                              viewBox="0 0 24 24"
                              fill="none"
                            >
                              <path
                                d="M5 12H19M19 12L12 5M19 12L12 19"
                                stroke="currentColor"
                                strokeWidth="2"
                                strokeLinecap="round"
                                strokeLinejoin="round"
                              />
                            </svg>
                          </>
                        )}
                      </button>
                    </div>
                  </div>
                </form>

                {/* Preview */}
                <div className="mt-4 rounded-2xl border border-dashed border-black/10 px-4 py-3 dark:border-white/10">
                  <p className="text-[10px] font-semibold uppercase tracking-wider text-black/30 dark:text-white/25">
                    Your temporary address
                  </p>

                  <p className="mt-1 break-all font-mono text-sm font-semibold text-black/70 dark:text-white/70">
                    {emailAddress}
                  </p>
                </div>
              </div>

              <div className="grid border-t border-black/5 bg-black/[0.015] sm:grid-cols-3 dark:border-white/5 dark:bg-white/[0.015]">
                <div className="border-b border-black/5 px-5 py-4 sm:border-b-0 sm:border-r dark:border-white/5">
                  <p className="text-xs font-semibold">Instant</p>
                  <p className="mt-1 text-[11px] leading-4 text-black/40 dark:text-white/30">
                    No account or password required.
                  </p>
                </div>

                <div className="border-b border-black/5 px-5 py-4 sm:border-b-0 sm:border-r dark:border-white/5">
                  <p className="text-xs font-semibold">Disposable</p>
                  <p className="mt-1 text-[11px] leading-4 text-black/40 dark:text-white/30">
                    Messages are automatically cleaned up.
                  </p>
                </div>

                <div className="px-5 py-4">
                  <p className="text-xs font-semibold">Simple</p>
                  <p className="mt-1 text-[11px] leading-4 text-black/40 dark:text-white/30">
                    Use it for unwanted subscriptions.
                  </p>
                </div>
              </div>
            </div>
          </section>

          {/* Warning */}
          <section className="mx-auto mt-8 max-w-2xl">
            <div className="flex gap-3 rounded-2xl border border-amber-500/15 bg-amber-500/[0.04] p-4 dark:border-amber-400/10 dark:bg-amber-400/[0.03]">
              <div className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400">
                <svg
                  width="15"
                  height="15"
                  viewBox="0 0 24 24"
                  fill="none"
                >
                  <path
                    d="M12 9V13M12 17H12.01M10.3 4.6L2.7 18C2 19.3 2.9 21 4.4 21H19.6C21.1 21 22 19.3 21.3 18L13.7 4.6C13 3.3 11 3.3 10.3 4.6Z"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </svg>
              </div>

              <div>
                <p className="text-xs font-semibold">
                  Don't use this for sensitive information
                </p>
                <p className="mt-1 text-xs leading-5 text-black/45 dark:text-white/35">
                  Temporary mailboxes are public and disposable. Avoid
                  passwords, personal information, financial data or
                  anything you need to keep private.
                </p>
              </div>
            </div>
          </section>
        </div>
      </main>
    </div>
  );
}

