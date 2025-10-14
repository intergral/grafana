import React, { createContext, useContext, useEffect, useRef, PropsWithChildren } from 'react';

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

  return (
    <OpsPilotBroadcastContext.Provider value={{ channel: channelRef.current }}>
      {children}
    </OpsPilotBroadcastContext.Provider>
  );
};
