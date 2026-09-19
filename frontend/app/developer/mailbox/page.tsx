"use client";

import { useState } from "react";
import type { SyntheticEvent } from "react";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { trackEvent } from "@/lib/posthog";

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE || "http://localhost:3001";

const MAIL_DOMAIN =
  process.env.NEXT_PUBLIC_MAIL_DOMAIN || "temp.mail.at";

type ExpiryOption = "1d" | "1w" | "1m" | "custom";

export default function Mailbox() {
  const router = useRouter();

  const [username, setUsername] = useState("");
  const [existingEmail, setExistingEmail] = useState("");
  const [expiry, setExpiry] = useState<ExpiryOption>("1d");
  const [customExpiry, setCustomExpiry] = useState("");

  const [isCreating, setIsCreating] = useState(false);
  const [isCheckingMailbox, setIsCheckingMailbox] = useState(false);

  const validateUsername = (input: string): string => {
    const beforeAtSign = input.split("@")[0];

    return beforeAtSign
      .toLowerCase()
      .replace(/[^a-z0-9]/g, "")
      .slice(0, 32);
  };

  const createMailbox = async (
    requestedUsername = username.trim()
  ) => {
    const cleanUsername = requestedUsername.trim();

    if (!cleanUsername) {
      toast.error("Enter a username first");
      return;
    }

    if (expiry === "custom" && !customExpiry) {
      toast.error("Choose an expiry date");
      return;
    }

    setIsCreating(true);

    try {
      trackEvent("mailbox_creation_attempt", {
        username: cleanUsername,
        method: "enter_key",
      });

      const response = await fetch(
        `${API_BASE}/api/mailboxes/custom`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            username: cleanUsername,

            // TODO: Send expiry information to the API
            // expiry,
            // expiresAt: expiry === "custom" ? customExpiry : undefined,
          }),
        }
      );

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

      const mailboxAddress =
        data.address || `${cleanUsername}@${MAIL_DOMAIN}`;

      const mailboxUsername = mailboxAddress.split("@")[0];

      router.push(
        `/mailbox/${encodeURIComponent(mailboxUsername)}`
      );
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

  const openExistingMailbox = async (
    e: SyntheticEvent<HTMLFormElement>
  ) => {
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
      // TODO: Replace/adjust API when mailbox lookup endpoint is finalized.
      const response = await fetch(
        `${API_BASE}/api/mailboxes/${encodeURIComponent(
          emailAddress
        )}/message`,
        {
          method: "POST",
        }
      );

      if (response.ok) {
        router.push(
          `/mailbox/${encodeURIComponent(emailUsername)}`
        );
        return;
      }

      if (response.status === 404) {
        const createUsername = validateUsername(emailUsername);

        setUsername(createUsername);

        toast.error("Mailbox not found", {
          description:
            "Create this temporary mailbox to start using it.",
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
    <div className="min-h-screen bg-[#f7f7f5] text-[#111] dark:bg-[#0b0c0c] dark:text-white">
      <main className="mx-auto flex min-h-screen max-w-5xl items-center px-5 py-12 lg:px-8">
        <div className="mx-auto w-full max-w-2xl">

          {/* Header */}
          <header className="mb-10">
            <div className="mb-5 flex items-center gap-2.5">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-black text-white dark:bg-white dark:text-black">
                <span className="text-sm font-bold">@</span>
              </div>

              <span className="text-sm font-semibold tracking-tight">
                {MAIL_DOMAIN}
              </span>
            </div>

            <h1 className="max-w-xl font-heading text-4xl font-semibold tracking-[-0.04em] sm:text-5xl">
              Temporary email,
              <br />
              <span className="text-black/35 dark:text-white/30">
                without the clutter.
              </span>
            </h1>

            <p className="mt-4 max-w-lg text-[15px] leading-6 text-black/50 dark:text-white/40">
              Create a disposable mailbox and choose how long it
              should remain available.
            </p>
          </header>

          {/* Create mailbox */}
          <section className="overflow-hidden rounded-2xl border border-black/[0.09] bg-white shadow-[0_12px_40px_rgba(0,0,0,0.05)] dark:border-white/[0.08] dark:bg-[#111313] dark:shadow-black/20">

            <div className="border-b border-black/[0.07] px-5 py-4 dark:border-white/[0.07]">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-semibold">
                    New mailbox
                  </p>

                  <p className="mt-0.5 text-xs text-black/40 dark:text-white/35">
                    Choose an address and expiration.
                  </p>
                </div>

                <span className="rounded-md bg-black/[0.04] px-2 py-1 font-mono text-[11px] text-black/45 dark:bg-white/[0.05] dark:text-white/40">
                  TEMP
                </span>
              </div>
            </div>

            <form onSubmit={handleSubmit} className="p-5 sm:p-6">

              {/* Username */}
              <div>
                <label
                  htmlFor="mail"
                  className="mb-2 block text-xs font-medium text-black/50 dark:text-white/40"
                >
                  Email address
                </label>

                <div className="flex h-12 overflow-hidden rounded-xl border border-black/[0.10] bg-[#fafaf9] transition focus-within:border-black/25 focus-within:ring-4 focus-within:ring-black/[0.035] dark:border-white/[0.10] dark:bg-[#0b0c0c] dark:focus-within:border-white/25 dark:focus-within:ring-white/[0.035]">

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
                    className="min-w-0 flex-1 bg-transparent px-3.5 text-[15px] font-medium outline-none placeholder:text-black/25 disabled:opacity-50 dark:placeholder:text-white/20"
                  />

                  <div className="flex items-center border-l border-black/[0.07] px-3 text-sm text-black/35 dark:border-white/[0.07] dark:text-white/30">
                    @{MAIL_DOMAIN}
                  </div>
                </div>
              </div>

              {/* Expiry */}
              <div className="mt-6">
                <div className="mb-2 flex items-center justify-between">
                  <label className="text-xs font-medium text-black/50 dark:text-white/40">
                    Expires after
                  </label>

                  {expiry === "custom" && customExpiry && (
                    <span className="font-mono text-[11px] text-black/40 dark:text-white/30">
                      {customExpiry}
                    </span>
                  )}
                </div>

                <div className="grid grid-cols-4 gap-1 rounded-xl bg-black/[0.035] p-1 dark:bg-white/[0.045]">
                  {[
                    { value: "1d", label: "1 day" },
                    { value: "1w", label: "1 week" },
                    { value: "1m", label: "1 month" },
                    { value: "custom", label: "Custom" },
                  ].map((option) => {
                    const selected = expiry === option.value;

                    return (
                      <button
                        key={option.value}
                        type="button"
                        onClick={() =>
                          setExpiry(
                            option.value as ExpiryOption
                          )
                        }
                        className={`h-9 rounded-lg text-xs font-medium transition ${
                          selected
                            ? "bg-white text-black shadow-sm dark:bg-[#1d2020] dark:text-white"
                            : "text-black/45 hover:text-black dark:text-white/40 dark:hover:text-white"
                        }`}
                      >
                        {option.label}
                      </button>
                    );
                  })}
                </div>

                {expiry === "custom" && (
                  <div className="mt-2">
                    <input
                      type="date"
                      value={customExpiry}
                      min={new Date()
                        .toISOString()
                        .split("T")[0]}
                      onChange={(e) =>
                        setCustomExpiry(e.target.value)
                      }
                      className="h-10 w-full rounded-lg border border-black/[0.09] bg-[#fafaf9] px-3 text-sm outline-none transition focus:border-black/20 focus:ring-4 focus:ring-black/[0.035] dark:border-white/[0.09] dark:bg-[#0b0c0c] dark:focus:border-white/20"
                    />
                  </div>
                )}
              </div>

              {/* Preview */}
              <div className="mt-5 flex items-center justify-between rounded-xl bg-black/[0.025] px-3.5 py-3 dark:bg-white/[0.035]">
                <span className="text-xs text-black/40 dark:text-white/35">
                  Your address
                </span>

                <span className="max-w-[65%] truncate font-mono text-xs font-medium text-black/60 dark:text-white/55">
                  {emailAddress}
                </span>
              </div>

              {/* Create */}
              <button
                type="submit"
                disabled={
                  isCreating ||
                  isCheckingMailbox ||
                  !username.trim()
                }
                className="mt-4 h-11 w-full rounded-xl bg-black text-sm font-semibold text-white transition hover:bg-black/80 active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-30 dark:bg-white dark:text-black dark:hover:bg-white/85"
              >
                {isCreating ? "Creating mailbox..." : "Create mailbox"}
              </button>
            </form>
          </section>

          {/* Existing mailbox */}
          <section className="mt-8">
            <div className="mb-3 flex items-center gap-3">
              <div className="h-px flex-1 bg-black/[0.07] dark:bg-white/[0.07]" />

              <span className="text-[11px] font-medium uppercase tracking-[0.12em] text-black/30 dark:text-white/25">
                Existing mailbox
              </span>

              <div className="h-px flex-1 bg-black/[0.07] dark:bg-white/[0.07]" />
            </div>

            <form
              onSubmit={openExistingMailbox}
              className="flex flex-col gap-2 sm:flex-row"
            >
              <input
                type="text"
                value={existingEmail}
                autoComplete="off"
                placeholder={`yourname@${MAIL_DOMAIN}`}
                disabled={isCheckingMailbox}
                onChange={(e) =>
                  setExistingEmail(e.target.value)
                }
                className="h-11 min-w-0 flex-1 rounded-xl border border-black/[0.09] bg-white px-3.5 text-sm outline-none transition placeholder:text-black/25 focus:border-black/20 focus:ring-4 focus:ring-black/[0.035] dark:border-white/[0.09] dark:bg-[#111313] dark:placeholder:text-white/20 dark:focus:border-white/20 dark:focus:ring-white/[0.035]"
              />

              <button
                type="submit"
                disabled={isCheckingMailbox}
                className="h-11 rounded-xl border border-black/[0.09] bg-white px-5 text-sm font-medium transition hover:bg-black/[0.025] active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-50 dark:border-white/[0.09] dark:bg-[#111313] dark:hover:bg-white/[0.04]"
              >
                {isCheckingMailbox
                  ? "Checking..."
                  : "Open mailbox"}
              </button>
            </form>
          </section>

          {/* Minimal footer */}
          <footer className="mt-10 flex items-center justify-between text-[11px] text-black/30 dark:text-white/25">
            <span>Disposable email service</span>

            <span className="font-mono">
              {MAIL_DOMAIN}
            </span>
          </footer>
        </div>
      </main>
    </div>
  );
}