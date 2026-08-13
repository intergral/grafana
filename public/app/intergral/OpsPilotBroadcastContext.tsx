import React, { createContext, useContext, useEffect, useRef, PropsWithChildren } from 'react';

import { useIframeNavigation } from './useIframeNavigation';
import { useOpspilotMetadata } from './useOpspilotMetadata';

interface OpsPilotBroadcastContextValue {
  channel: BroadcastChannel | null;
}

const OpsPilotBroadcastContext = createContext<OpsPilotBroadcastContextValue>({ channel: null });

export const useOpsPilotBroadcast = () => useContext(OpsPilotBroadcastContext);

export const OpsPilotBroadcastProvider: React.FC<PropsWithChildren> = ({ children }) => {
  const channelRef = useRef<BroadcastChannel | null>(null);

  useEffect(() => {
    // Create broadcast channel
    channelRef.current = new BroadcastChannel('opspilot');

    // Cleanup on unmount
    return () => {
      channelRef.current?.close();
    };
  }, []);

  // Enable iframe navigation from parent window
  useIframeNavigation();

  // Handle OpsPilot metadata requests
  useOpspilotMetadata();

  return (
    <OpsPilotBroadcastContext.Provider value={{ channel: channelRef.current }}>
      {children}
    </OpsPilotBroadcastContext.Provider>
  );
};
