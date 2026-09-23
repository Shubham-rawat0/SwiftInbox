"use client";

import { useState } from "react";
import type { SyntheticEvent } from "react";
import { toast } from "sonner";
import { useRouter } from "next/navigation";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:3001";
const MAIL_DOMAIN = process.env.NEXT_PUBLIC_MAIL_DOMAIN || "temp.mail.at";

export default function Home() {
  const router = useRouter();
  const [username, setUsername] = useState("");
  const [existingEmail, setExistingEmail] = useState("");
  const [isCreating, setIsCreating] = useState(false);
  const [isCheckingMailbox, setIsCheckingMailbox] = useState(false);

  const validateUsername = (input: string): string => {
    const beforeAtSign = input.split("@")[0];

    return beforeAtSign
      .toLowerCase()
      .replace(/[^a-z0-9]/g, "")
      .slice(0, 32);
  };

  const createMailbox = async (requestedUsername = username.trim()) => {
    const cleanUsername = requestedUsername.trim();

    if (!cleanUsername) {
      toast.error("Enter a username first");
      return;
    }

    setIsCreating(true);

    try {
      const response = await fetch(`${API_BASE}/api/mailboxes/custom`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          username: cleanUsername,
        }),
      });

      const data = await response.json().catch(() => ({}));

      if (response.status === 409) {
        toast.error("Username already taken", {
          description:
            "This address already exists. Try another username or open your existing mailbox below.",
        });
        return;
      }

      if (response.status === 429) {
        toast.error("Rate limit exceeded", {
          description:
            data.error ||
            "Too many mailboxes created. Please try again after 1 hour.",
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

      toast.success("Mailbox created");

      const mailboxAddress = data.address || `${cleanUsername}@${MAIL_DOMAIN}`;
      const mailboxUsername = mailboxAddress.split("@")[0];
      router.push(`/mailbox/${encodeURIComponent(mailboxUsername)}`);
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

  const openExistingMailbox = async (e: SyntheticEvent<HTMLFormElement>) => {
    e.preventDefault();

    const cleanEmail = existingEmail.trim().toLowerCase();

    if (!cleanEmail) {
      toast.error("Enter your email address");
      return;
    }

    const emailAddress = cleanEmail.includes("@")
      ? cleanEmail
      : `${cleanEmail}@${MAIL_DOMAIN}`;
    const emailUsername = emailAddress.split("@")[0];

    if (!emailUsername) {
      toast.error("Enter a valid email address");
      return;
    }

    setIsCheckingMailbox(true);

    try {
      const response = await fetch(
        `${API_BASE}/api/mailboxes/${encodeURIComponent(emailAddress)}/message`,
        { method: "POST" }
      );

      if (response.ok) {
        router.push(`/mailbox/${encodeURIComponent(emailUsername)}`);
        return;
      }

      if (response.status === 404) {
        const createUsername = validateUsername(emailUsername);
        setUsername(createUsername);
        toast.error("Mailbox not found", {
          description: "Create this temporary mailbox to start using it.",
          action: {
            label: "Create mailbox",
            onClick: () => createMailbox(createUsername),
          },
        });
        return;
      }

      toast.error("Could not check mailbox", {
        description: "Please try again in a moment.",
      });
    } catch (error) {
      console.error("Error checking mailbox:", error);
      toast.error("Connection failed", {
        description: "Could not connect to the mail server.",
      });
    } finally {
      setIsCheckingMailbox(false);
    }
  };

  const emailAddress = `${username || "username"}@${MAIL_DOMAIN}`;

  return (
    <div
      className="min-h-screen bg-[#f7f6f3] text-[#14161a] dark:bg-[#0d0e0c] dark:text-[#f3f2ef]"
      style={{ fontFamily: "var(--font-sans)" }}
    >
      <style jsx global>{`
        @import url("https://fonts.googleapis.com/css2?family=Archivo:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap");
        :root {
          --font-sans: "Archivo", ui-sans-serif, system-ui, -apple-system,
            sans-serif;
          --font-mono: "JetBrains Mono", ui-monospace, SFMono-Regular, Menlo,
            monospace;
        }
      `}</style>

      {/* Top bar */}
      <header className="mx-auto flex max-w-[1240px] items-center justify-between px-6 py-6 sm:px-10 lg:px-16">
        <span
          className="text-[15px] font-semibold tracking-tight"
          style={{ fontFamily: "var(--font-mono)" }}
        >
          {MAIL_DOMAIN}
        </span>

        <span className="hidden items-center gap-2 text-[13px] text-black/40 dark:text-white/35 sm:flex">
          <span className="h-1.5 w-1.5 rounded-full bg-[#4a6a52]" />
          Mailboxes live now
        </span>
      </header>

      <main className="mx-auto max-w-[1240px] px-6 pb-24 sm:px-10 lg:px-16">
        <section className="grid gap-14 pt-6 lg:grid-cols-[1.05fr_1fr] lg:gap-20 lg:pt-14">
          {/* Left: narrative */}
          <div className="max-w-[48ch]">
            <h1 className="text-[38px] font-semibold leading-[1.1] tracking-[-0.02em] sm:text-[50px] lg:text-[56px]">
              An inbox you don&apos;t have to keep.
            </h1>

            <p className="mt-5 text-[17px] leading-7 text-black/60 dark:text-white/55">
              Spin up a disposable address for sign-ups, downloads and
              one-off verifications. No account, no cleanup — it closes
              itself.
            </p>

            <dl className="mt-10 space-y-5 border-t border-black/[0.08] pt-7 dark:border-white/[0.08]">
              <div className="flex gap-5">
                <dt
                  className="w-28 shrink-0 pt-0.5 text-[13px] text-black/40 dark:text-white/35"
                  style={{ fontFamily: "var(--font-mono)" }}
                >
                  No account
                </dt>
                <dd className="text-[15px] leading-6 text-black/65 dark:text-white/55">
                  Pick a username below and you&apos;re already in the
                  mailbox.
                </dd>
              </div>

              <div className="flex gap-5">
                <dt
                  className="w-28 shrink-0 pt-0.5 text-[13px] text-black/40 dark:text-white/35"
                  style={{ fontFamily: "var(--font-mono)" }}
                >
                  Self-clearing
                </dt>
                <dd className="text-[15px] leading-6 text-black/65 dark:text-white/55">
                  Messages are removed automatically — nothing to delete by
                  hand.
                </dd>
              </div>

              <div className="flex gap-5">
                <dt
                  className="w-28 shrink-0 pt-0.5 text-[13px] text-black/40 dark:text-white/35"
                  style={{ fontFamily: "var(--font-mono)" }}
                >
                  24 hours
                </dt>
                <dd className="text-[15px] leading-6 text-black/65 dark:text-white/55">
                  Every mailbox expires a day after it&apos;s created.
                </dd>
              </div>
            </dl>
          </div>

          {/* Right: functional panel */}
          <div className="rounded-2xl border border-black/[0.09] bg-white dark:border-white/[0.09] dark:bg-[#131413]">
            <div className="p-6 sm:p-7">
              <p className="text-[15px] font-semibold">Create a mailbox</p>
              <p className="mt-1 text-[14px] leading-6 text-black/45 dark:text-white/35">
                Choose a username for your new address.
              </p>

              <form onSubmit={handleSubmit} className="mt-5">
                <div className="flex h-[52px] items-center rounded-xl border border-black/[0.10] bg-[#f8f8f6] pl-4 pr-1.5 transition-colors focus-within:border-black/25 dark:border-white/[0.10] dark:bg-[#0d0e0c] dark:focus-within:border-white/25">
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
                    className="min-w-0 flex-1 bg-transparent text-[16px] outline-none placeholder:text-black/25 disabled:opacity-50 dark:placeholder:text-white/20"
                    style={{ fontFamily: "var(--font-mono)" }}
                  />

                  <span
                    className="hidden shrink-0 pr-2 text-[14px] text-black/35 sm:block dark:text-white/30"
                    style={{ fontFamily: "var(--font-mono)" }}
                  >
                    @{MAIL_DOMAIN}
                  </span>

                  <button
                    type="submit"
                    disabled={isCreating || isCheckingMailbox || !username.trim()}
                    className="h-10 shrink-0 rounded-lg bg-[#14161a] px-4 text-[14px] font-medium text-white transition hover:bg-[#14161a]/85 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-30 dark:bg-white dark:text-black dark:hover:bg-white/85"
                  >
                    {isCreating ? "Creating…" : "Create"}
                  </button>
                </div>
              </form>

              <div className="mt-4 flex items-center justify-between gap-3 rounded-lg bg-black/[0.03] px-4 py-3 dark:bg-white/[0.04]">
                <span className="text-[13px] text-black/45 dark:text-white/35">
                  Your address
                </span>
                <span
                  className="truncate text-[13px] font-medium text-black/70 dark:text-white/70"
                  style={{ fontFamily: "var(--font-mono)" }}
                >
                  {emailAddress}
                </span>
              </div>
            </div>

            <div className="border-t border-black/[0.08] p-6 sm:p-7 dark:border-white/[0.08]">
              <p className="text-[15px] font-semibold">
                Already have an address?
              </p>

              <form
                onSubmit={openExistingMailbox}
                className="mt-4 flex flex-col gap-2.5 sm:flex-row"
              >
                <input
                  type="text"
                  value={existingEmail}
                  autoComplete="off"
                  placeholder={`yourname@${MAIL_DOMAIN}`}
                  disabled={isCheckingMailbox}
                  onChange={(e) => setExistingEmail(e.target.value)}
                  className="h-11 min-w-0 flex-1 rounded-lg border border-black/[0.10] bg-[#f8f8f6] px-3.5 text-[15px] outline-none transition focus:border-black/25 dark:border-white/[0.10] dark:bg-[#0d0e0c] dark:focus:border-white/25"
                  style={{ fontFamily: "var(--font-mono)" }}
                />

                <button
                  type="submit"
                  disabled={isCheckingMailbox}
                  className="h-11 shrink-0 rounded-lg border border-black/10 px-5 text-[14px] font-medium transition hover:bg-black/[0.03] active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50 dark:border-white/10 dark:hover:bg-white/[0.06]"
                >
                  {isCheckingMailbox ? "Checking…" : "Open"}
                </button>
              </form>
            </div>
          </div>
        </section>

        {/* Warning */}
        <section className="mt-16 border-l-2 border-[#b5762c]/50 pl-5">
          <p className="text-[14px] font-medium text-[#95601f] dark:text-[#d99a52]">
            Mailboxes are public
          </p>
          <p className="mt-1 max-w-[62ch] text-[14px] leading-6 text-black/55 dark:text-white/40">
            Anyone who knows or guesses the address can read what lands in
            it. Don&apos;t use a temporary mailbox for passwords, financial
            details or anything you wouldn&apos;t want a stranger to see.
          </p>
        </section>
      </main>
    </div>
  );
}