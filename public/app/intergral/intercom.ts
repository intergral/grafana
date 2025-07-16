import { useEffect } from 'react';

// Declare Intercom as a global function
declare global {
  interface Window {
    Intercom: any;
  }
}

export function useIntercom(userName: string, userEmail: string) {
  useEffect(() => {
    // Check if we're in a browser environment
    if (typeof window === 'undefined' || typeof document === 'undefined') {
      console.warn('Intercom setup aborted: Not in a browser environment');
      return;
    }

    // Intercom setup function
    const setupIntercom = () => {
      (function() {
        let w = window as any;
        let ic = w.Intercom;
        if (typeof ic === "function") {
          ic('reattach_activator');
          ic('update', w.intercomSettings);
        } else {
          let d = document;
          let i = function () {
            (i as any).c(arguments);
          };
          (i as any).q = [];
          (i as any).c = function(args: any) {
            (i as any).q.push(args);
          };
          w.Intercom = i;
          let l = function () {
            let s = d.createElement('script');
            s.type = 'text/javascript';
            s.async = true;
            s.src = 'https://widget.intercom.io/widget/ok1wowgi';
            let x = d.getElementsByTagName('script')[0];
            let parent = x?.parentNode || document.body
            parent.insertBefore(s, x || null);
          };
          if (document.readyState === 'complete') {
            l();
          } else if (w.attachEvent) {
            w.attachEvent('onload', l);
          } else {
            w.addEventListener('load', l, false);
          }
        }
      })();

      window.Intercom("boot", {
        api_base: "https://api-iam.intercom.io",
        app_id: "ok1wowgi",
        name: userName,
        email: userEmail,
      });
    };

    setupIntercom();

    // Cleanup function
    return () => {
      if (window.Intercom) {
        window.Intercom('shutdown');
      }
    };
  }, [userName, userEmail]); // Re-run if userName or userEmail changes
}
