
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
      const response = await fetch(
        `${API_BASE}/api/mailboxes/custom`,
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
            "This address already exists. Try another username or open your existing mailbox below.",
          duration: 3000,
        });
        return;
      }

      if (response.status === 429) {
        toast.error("Rate limit exceeded", {
          description:
            data.error ||
            "Too many mailboxes created. Please try again after 1 hour.",
          duration: 3000,
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
          duration: 3000,
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
    <div className="min-h-screen bg-[#f7f7f5] text-[#111] dark:bg-[#0b0c0c] dark:text-white">

      {/* Main */}
      <main className="mx-auto max-w-6xl px-5 py-6 lg:px-8 lg:py-5">
        <div className="w-full">
          {/* Hero */}
          <section className="mx-auto max-w-2xl text-center">
            <h1 className="font-heading text-[48px] font-bold leading-[1.05] tracking-[-0.04em] sm:text-[64px]">
              Temporary email.
              <br />
              <span className="text-black/35 dark:text-white/30">
                Simple and private.
              </span>
            </h1>

            <p className="mx-auto mt-6 max-w-xl text-[17px] leading-7 text-black/55 dark:text-white/45 sm:text-[18px]">
              Create a disposable email address in seconds. Use it for
              sign-ups, newsletters and services you don&apos;t want connected
              to your personal inbox.
            </p>
          </section>

          {/* Create mailbox */}
          <section className="mx-auto mt-10 max-w-xl">
            <div className="rounded-[26px] border border-black/[0.08] bg-white shadow-[0_18px_60px_rgba(0,0,0,0.07)] dark:border-white/[0.08] dark:bg-[#111313] dark:shadow-black/20">
              <div className="p-5 sm:p-7">
                <div className="mb-5">
                  <p className="font-heading text-[17px] font-semibold tracking-[-0.01em]">
                    Create a temporary address
                  </p>

                  <p className="mt-1 text-[15px] leading-6 text-black/45 dark:text-white/35">
                    Choose a username for your new mailbox.
                  </p>
                </div>

                <form onSubmit={handleSubmit}>
                  <div className="rounded-2xl border border-black/[0.09] bg-[#f8f8f7] p-1.5 transition-all focus-within:border-black/20 focus-within:ring-4 focus-within:ring-black/[0.035] dark:border-white/[0.09] dark:bg-[#0b0c0c] dark:focus-within:border-white/20 dark:focus-within:ring-white/[0.035]">
                    <div className="flex h-[52px] items-center">
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
                        className="min-w-0 flex-1 bg-transparent px-4 text-[17px] font-medium outline-none placeholder:text-black/25 disabled:opacity-50 dark:placeholder:text-white/20"
                      />

                      <span className="hidden shrink-0 pr-3 text-[15px] text-black/35 sm:block dark:text-white/30">
                        @{MAIL_DOMAIN}
                      </span>

                      <button
                        type="submit"
                        disabled={isCreating || isCheckingMailbox || !username.trim()}
                        className="h-[44px] shrink-0 rounded-xl bg-black px-5 text-[14px] font-semibold text-white transition hover:bg-black/80 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-30 dark:bg-white dark:text-black dark:hover:bg-white/85"
                      >
                        {isCreating ? "Creating..." : "Create mailbox"}
                      </button>
                    </div>
                  </div>
                </form>

                {/* Address preview */}
                <div className="mt-4 rounded-xl bg-black/[0.025] px-4 py-3 dark:bg-white/[0.035]">
                  <p className="text-[12px] font-medium uppercase tracking-[0.08em] text-black/30 dark:text-white/25">
                    Your address
                  </p>

                  <p className="mt-1.5 break-all font-mono text-[15px] font-medium text-black/65 dark:text-white/60">
                    {emailAddress}
                  </p>
                </div>
              </div>
            </div>
          </section>

          {/* Existing mailbox */}
          <section className="mx-auto mt-5 max-w-xl">
            <div className="rounded-[22px] border border-black/[0.07] bg-white/70 p-5 dark:border-white/[0.07] dark:bg-white/[0.025] sm:p-6">
              <div className="mb-4">
                  <p className="font-heading text-[17px] font-semibold">
                  Already have a mailbox?
                </p>

                <p className="mt-1 text-[15px] leading-6 text-black/45 dark:text-white/35">
                  Enter your existing temporary email address to open your
                  inbox.
                </p>
              </div>

              <form
                onSubmit={openExistingMailbox}
                className="flex flex-col gap-2.5 sm:flex-row"
              >
                <input
                  type="text"
                  value={existingEmail}
                  autoComplete="off"
                  placeholder={`yourname@${MAIL_DOMAIN}`}
                  disabled={isCheckingMailbox}
                  onChange={(e) => setExistingEmail(e.target.value)}
                  className="h-12 min-w-0 flex-1 rounded-xl border border-black/[0.09] bg-[#f8f8f7] px-3.5 text-[16px] outline-none transition focus:border-black/20 focus:ring-4 focus:ring-black/[0.035] dark:border-white/[0.09] dark:bg-[#0b0c0c] dark:focus:border-white/20 dark:focus:ring-white/[0.035]"
                />

                <button
                  type="submit"
                  disabled={isCheckingMailbox}
                  className="h-12 rounded-xl border border-black/10 bg-white px-5 text-[14px] font-semibold transition hover:bg-black/[0.03] active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50 dark:border-white/10 dark:bg-white/[0.04] dark:hover:bg-white/[0.07]"
                >
                  {isCheckingMailbox ? "Checking..." : "Open mailbox"}
                </button>
              </form>
            </div>
          </section>

          {/* Features */}
          <section className="mx-auto mt-10 grid max-w-xl gap-3 sm:grid-cols-3">
            <div className="rounded-2xl border border-black/[0.06] bg-white/50 p-4 dark:border-white/[0.06] dark:bg-white/[0.02]">
              <p className="text-[15px] font-semibold">Instant</p>
              <p className="mt-1.5 text-[14px] leading-5 text-black/45 dark:text-white/35">
                Create an inbox without an account.
              </p>
            </div>

            <div className="rounded-2xl border border-black/[0.06] bg-white/50 p-4 dark:border-white/[0.06] dark:bg-white/[0.02]">
              <p className="text-[15px] font-semibold">Disposable</p>
              <p className="mt-1.5 text-[14px] leading-5 text-black/45 dark:text-white/35">
                Messages are automatically removed.
              </p>
            </div>

            <div className="rounded-2xl border border-black/[0.06] bg-white/50 p-4 dark:border-white/[0.06] dark:bg-white/[0.02]">
              <p className="text-[15px] font-semibold">24 hours</p>
              <p className="mt-1.5 text-[14px] leading-5 text-black/45 dark:text-white/35">
                Mailboxes expire after one day.
              </p>
            </div>
          </section>

          {/* Warning */}
          <section className="mx-auto mt-8 mb-10 max-w-xl">
            <div className="rounded-2xl border border-amber-500/[0.12] bg-amber-500/[0.035] px-4 py-3.5 dark:border-amber-400/[0.10] dark:bg-amber-400/[0.025]">
              <p className="text-[14px] font-semibold text-amber-700 dark:text-amber-400">
                Important
              </p>

              <p className="mt-1 text-[14px] leading-6 text-black/50 dark:text-white/40">
                Temporary mailboxes are public. Do not use them for
                passwords, personal information, financial data or anything
                sensitive.
              </p>
            </div>
          </section>
        </div>
      </main>
    </div>
  );
}

