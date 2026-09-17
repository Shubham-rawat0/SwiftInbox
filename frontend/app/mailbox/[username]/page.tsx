"use client"
import { useParams } from "next/navigation"
import { useState, useEffect, useLayoutEffect, useCallback, useRef } from "react"
import { Button } from "@/components/ui/button"
import { RefreshCw as Refresh, Lock, ArrowRightLeft,Copy, Check  } from "lucide-react"
import Link from "next/link"
import { toast } from "sonner"
import { fetchMessages } from "@/lib/api"
import { trackEvent } from "@/lib/posthog"

const MAIL_DOMAIN = process.env.NEXT_PUBLIC_MAIL_DOMAIN || "temp.mail.at"

export default function MailboxPage() {
  const params = useParams()
  const username = params.username as string
  const address = username.includes("@") ? username : `${username}@${MAIL_DOMAIN}`

  const [emails, setEmails] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [isListening, setIsListening] = useState(true)
  const [lastChecked, setLastChecked] = useState(new Date())
  const [failedAttempts, setFailedAttempts] = useState(0)
  const [apiErrorState, setApiErrorState] = useState(false)
  const [hasStableEmails, setHasStableEmails] = useState(false)
  const [unchangedPolls, setUnchangedPolls] = useState(0)
  const retryTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const lastEmailSignatureRef = useRef("")
  const requestInFlightRef = useRef(false)

  useLayoutEffect(() => {
    const resetScroll = () => {
      window.scrollTo({ top: 0, left: 0, behavior: "auto" });
      document.documentElement.scrollTop = 0;
      document.body.scrollTop = 0;
    };

    resetScroll();
    const frame = window.requestAnimationFrame(resetScroll);

    return () => window.cancelAnimationFrame(frame);
  }, [username]);

  const loadEmails = async (forceRefresh = false) => {
    if ((apiErrorState && failedAttempts >= 5) || requestInFlightRef.current) {
      return;
    }

    requestInFlightRef.current = true;
    setLoading(true);
    
    try {
      const result = await fetchMessages(address, forceRefresh);
      
      const newEmails = result.messages || [];
      setEmails(newEmails);
      
      const emailSignature = newEmails.map(email => email.id).join(",");
      if (emailSignature && emailSignature === lastEmailSignatureRef.current) {
        // to increase polling from 45 sec to 2 min if email count doesn't change
        setUnchangedPolls(prev => prev + 1); 
      } else {
        setUnchangedPolls(0);
        setHasStableEmails(false);
      }
      lastEmailSignatureRef.current = emailSignature;
      
      if (failedAttempts > 0) {
        setFailedAttempts(0);
        setApiErrorState(false);
      }
    } catch (error) {
      console.error('Failed to load emails:', error);
      
      const newFailedAttempts = failedAttempts + 1;
      setFailedAttempts(newFailedAttempts);
      setApiErrorState(true);

      toast.error(`Cannot reach server right now. Retry ${newFailedAttempts}/5`, {
      });

      if (newFailedAttempts >= 5) {
        toast.error("Maximum retry attempts reached. Please refresh the page to try again.");
        setIsListening(false);
      }
    } finally {
      requestInFlightRef.current = false;
      setLoading(false);
      setLastChecked(new Date());
    }
  }

  useEffect(() => {
      if (unchangedPolls >= 3 && lastEmailSignatureRef.current) {
        setHasStableEmails(true)
      }
  }, [unchangedPolls])

  useEffect(() => {
     loadEmails(true)
  }, [])

  useEffect(() => {
      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current);
        retryTimeoutRef.current = null;
      }

      if (apiErrorState && failedAttempts >= 5) {
        return;
      }

      const getBackoffTime = () => {
        if (hasStableEmails && emails.length > 0) {
          return 120000; 
        }
        
        if (failedAttempts === 0) return 45000;
        
        return Math.min(Math.pow(2, failedAttempts) * 5000, 300000);
      };

      const interval = setInterval(() => {
        if (isListening) {
        loadEmails(true);
    }
      }, getBackoffTime())
      return () => clearInterval(interval)
  }, [isListening, failedAttempts, apiErrorState, hasStableEmails, emails.length])


  const manualRefresh = async () => {
    trackEvent('manual_refresh', {
      username: username,
      current_email_count: emails.length
    });

    setApiErrorState(false);
    setFailedAttempts(0);
    setHasStableEmails(false);
    setUnchangedPolls(0);
    setIsListening(true);

    toast("Refreshing mailbox...", {
    });

    setRefreshing(true);

    try {
      const result = await fetchMessages(address, true);
      setEmails(result.messages || []);

      trackEvent('manual_refresh_success', {
        username: username,
        new_email_count: result.messages?.length || 0,
        previous_email_count: emails.length
      });

      toast.success("Mailbox refreshed!", {
      });
    } catch (error) {
      const err = error as Error;
      trackEvent('manual_refresh_failed', {
        username: username,
        error: err.message
      });

      toast.error("Failed to refresh mailbox. Please try again later.", {
      });
    } finally {
      setRefreshing(false);
      setLastChecked(new Date());
    }
  }
  
const EmailAddressDisplay = () => {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    const email = `${username}@temp.mail.at`;

    try {
      await navigator.clipboard.writeText(email);

      trackEvent("email_copied", {
        username,
        method: "email_display_click",
        email_address: email,
      });

      setCopied(true);
      toast.success("Email copied to clipboard!");
      setTimeout(() => setCopied(false), 1800);
    } catch {
      toast.error("Failed to copy email");
    }
  };

  return (
    <button
      type="button"
      onClick={handleCopy}
      aria-label="Copy email address"
      title="Click to copy email address"
      className="
        group relative inline-flex items-center gap-3
        px-5 py-2.5 mb-4
        rounded-full
        bg-white border border-gray-200
        text-gray-900
        font-medium 
        shadow-[0_1px_2px_rgba(0,0,0,0.04)]
        transition-all duration-200 ease-out
        hover:border-gray-300 hover:shadow-[0_2px_8px_rgba(0,0,0,0.06)]
        active:scale-[0.98]
        cursor-pointer
        focus:outline-none focus:ring-2 focus:ring-gray-300 focus:ring-offset-2
      "
    >
      <span className="font-mono tracking-tight text-gray-800">
        {username}
        <span className="text-gray-400">@temp.mail.at</span>
      </span>

      <span className="relative flex h-[18px] w-[18px] items-center justify-center shrink-0">
        <Copy
          size={18}
          className={`
            absolute text-gray-400
            transition-all duration-200
            ${copied ? "opacity-0 scale-75" : "opacity-100 scale-100 group-hover:text-gray-600"}
          `}
        />
        <Check
          size={18}
          className={`
            absolute text-emerald-600
            transition-all duration-200
            ${copied ? "opacity-100 scale-100" : "opacity-0 scale-75"}
          `}
        />
      </span>
    </button>
  );
};
  
  const ActionButtons = () => (
   <div className="flex flex-col gap-2 px-2 sm:px-0">
  <div className="flex flex-row justify-center gap-2 sm:gap-3">
    <Button
      className="bg-gray-900 hover:bg-gray-800 text-white text-sm sm:text-base flex-1 sm:flex-initial sm:min-w-[120px] shadow-sm transition-colors"
      onClick={() => {
        const email = `${username}@temp.mail.at`;
        navigator.clipboard.writeText(email);
        trackEvent('email_copied', {
          username: username,
          method: 'copy_button',
          email_address: email
        });
        toast.success("Email copied to clipboard!");
      }}
    >
      <Copy className="w-4 h-4 mr-2" />
      Copy
    </Button>

    <Link href="/" className="flex-1 sm:flex-initial">
      <Button
        variant="outline"
        className="text-sm sm:text-base w-full sm:min-w-[140px] border-gray-300 text-gray-700 hover:bg-gray-50 hover:border-gray-400 transition-colors"
      >
        <ArrowRightLeft className="w-4 h-4 mr-2" />
        Change Email
      </Button>
    </Link>

    <Button
      variant="outline"
      onClick={manualRefresh}
      className="hidden sm:flex text-sm sm:text-base sm:min-w-[140px] border-gray-300 text-gray-700 hover:bg-gray-50 hover:border-gray-400 transition-colors"
    >
      <Refresh className="w-4 h-4 mr-2" />
      Refresh
    </Button>
  </div>

  <div className="flex justify-center sm:hidden">
    <Button
      variant="outline"
      onClick={manualRefresh}
      className="text-sm w-full max-w-[300px] border-gray-300 text-gray-700 hover:bg-gray-50 hover:border-gray-400 transition-colors"
    >
      <Refresh className="w-4 h-4 mr-2" />
      Refresh
    </Button>
  </div>
</div>
  );
  
  const MailboxHeader = () => (
    <div className="text-center mb-4 sm:mb-6">
      <EmailAddressDisplay />
      <ActionButtons />
    </div>
  );
  
const EmailsList = () => {
  // Manual refresh
  if (refreshing) {
    return (
      <div className="flex min-h-[50vh] flex-col items-center justify-center rounded-2xl border border-gray-200 bg-white px-6 py-12 text-center shadow-sm dark:border-gray-700 dark:bg-gray-800">
        <div className="flex h-14 w-14 items-center justify-center rounded-full bg-gray-100 dark:bg-gray-700">
          <Refresh className="h-6 w-6 animate-spin text-gray-500 dark:text-gray-300" />
        </div>

        <h3 className="mt-5 text-lg font-semibold text-gray-900 dark:text-white">
          Refreshing mailbox
        </h3>

        <p className="mt-1 max-w-sm text-sm text-gray-500 dark:text-gray-400">
          Checking for new messages...
        </p>
      </div>
    );
  }

  // API permanently failed
  if (apiErrorState && failedAttempts >= 5) {
    return (
      <div className="flex min-h-[50vh] flex-col items-center justify-center rounded-2xl border border-red-200 bg-red-50/50 px-6 py-12 text-center dark:border-red-900/60 dark:bg-red-950/20">
        <div className="flex h-14 w-14 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/40">
          <span className="text-2xl">!</span>
        </div>

        <h3 className="mt-5 text-lg font-semibold text-gray-900 dark:text-white">
          Unable to reach the server
        </h3>

        <p className="mt-2 max-w-md text-sm leading-6 text-gray-500 dark:text-gray-400">
          We couldn't connect to the mailbox server after several attempts.
          Your emails are safe. Try refreshing when you're ready.
        </p>

        <Button
          variant="outline"
          size="sm"
          onClick={manualRefresh}
          className="mt-6 gap-2 border-gray-300 bg-white dark:border-gray-600 dark:bg-gray-800"
        >
          <Refresh className="h-4 w-4" />
          Try again
        </Button>
      </div>
    );
  }

  // Empty mailbox
  if (emails.length === 0) {
    return (
      <div className="flex min-h-[50vh] flex-col items-center justify-center rounded-2xl border border-dashed border-gray-300 bg-gray-50/50 px-6 py-12 text-center dark:border-gray-600 dark:bg-gray-800/50">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-gray-100 dark:bg-gray-700">
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.7"
            className="h-8 w-8 text-gray-400 dark:text-gray-300"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M3 8.5 12 14l9-5.5M4.5 5h15A1.5 1.5 0 0 1 21 6.5v11a1.5 1.5 0 0 1-1.5 1.5h-15A1.5 1.5 0 0 1 3 17.5v-11A1.5 1.5 0 0 1 4.5 5Z"
            />
          </svg>
        </div>

        <h3 className="mt-5 text-lg font-semibold text-gray-900 dark:text-white">
          Your inbox is empty
        </h3>

        <p className="mt-2 max-w-sm text-sm leading-6 text-gray-500 dark:text-gray-400">
          Emails sent to{" "}
          <span className="font-medium text-gray-700 dark:text-gray-300">
            {username}@temp.mail.at
          </span>{" "}
          will appear here automatically.
        </p>

        {isListening && (
          <div className="mt-5 inline-flex items-center gap-2 rounded-full bg-gray-100 px-3 py-1.5 text-xs font-medium text-gray-600 dark:bg-gray-700 dark:text-gray-300">
            <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-green-500" />
            Listening for new emails
          </div>
        )}
      </div>
    );
  }

  // Email list
  return (
    <div className="space-y-3">
      {Array.from(
        { length: Math.ceil(emails.length / 2) },
        (_, i) => {
          const email1 = emails[i * 2];
          const email2 = emails[i * 2 + 1];

          const EmailCard = ({ email }: { email: any }) => (
            <Link
              href={`/mailbox/${username}/message/${email.id}`}
              className="group block h-full"
            >
              <article
                className="
                  h-full rounded-xl border border-gray-200
                  bg-white p-4 shadow-sm
                  transition-all duration-200
                  hover:-translate-y-0.5 hover:border-gray-300 hover:shadow-md
                  dark:border-gray-700 dark:bg-gray-800
                  dark:hover:border-gray-600
                "
              >
                <div className="flex items-start justify-between gap-4">
                  <h3 className="min-w-0 flex-1 truncate text-base font-semibold text-gray-900 dark:text-white">
                    {email.subject || "(No subject)"}
                  </h3>

                  <span className="shrink-0 text-xs text-gray-400 dark:text-gray-500">
                    {email.createdAt
                      ? new Date(email.createdAt).toLocaleDateString("en-US", {
                          month: "short",
                          day: "numeric",
                          year: "numeric",
                        })
                      : email.time}
                  </span>
                </div>

                <p className="mt-2 line-clamp-2 break-words text-sm leading-5 text-gray-500 dark:text-gray-400">
                  {email.preview || "No preview available"}
                </p>

                <div className="mt-4 flex items-center text-xs font-medium text-gray-400 transition-colors group-hover:text-gray-700 dark:text-gray-500 dark:group-hover:text-gray-300">
                  View message
                  <ArrowRightLeft className="ml-1.5 h-3.5 w-3.5" />
                </div>
              </article>
            </Link>
          );

          return (
            <div
              key={i}
              className="grid grid-cols-1 gap-3 md:grid-cols-2"
            >
              {email1 && <EmailCard email={email1} />}
              {email2 && <EmailCard email={email2} />}
            </div>
          );
        }
      )}
    </div>
  );
};


return (
  <>
  {/* mobile view */}
  <div className="md:hidden flex flex-col min-h-screen">
        <main className="flex-1 bg-white dark:bg-[#0D0E0E]">
          <div className="max-w-4xl mx-auto px-4 py-4 sm:px-6 lg:px-8 sm:py-8">
            <div className="border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg p-6 mb-6 sm:mb-8">
              <MailboxHeader />
            </div>
                
            <div className="space-y-3 sm:space-y-4">
              <EmailsList />
            </div>
          </div>
        </main>

      </div>

{/* desktop view */}
  <div className="hidden md:block font-sans antialiased">
    <div className="bg-white dark:bg-zinc-900/60 relative z-10 min-h-screen flex flex-col">

      <main className="flex-1">
        <div className="max-w-5xl mx-auto px-6 lg:px-8 py-10">

          {/* Mailbox Header */}
          <div className="border-2 border-zinc-200 dark:border-zinc-800 rounded-2xl bg-zinc-50 dark:bg-[#0A0B0B] p-6 mb-8 shadow-sm hover:shadow-md transition-shadow duration-300">
            <MailboxHeader />
          </div>


          {/* Emails */}
          <div className="space-y-3 border-2 border-zinc-200 dark:border-zinc-800 bg-zinc-50 dark:bg-[#0A0B0B] rounded-2xl p-4  shadow-sm">
            <EmailsList />
          </div>

        </div>
      </main>

    </div>
  </div>
  </>
)


}